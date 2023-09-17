package hooks

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	lop "github.com/samber/lo/parallel"

	"github.com/dirk/quickhook/internal"
	"github.com/dirk/quickhook/repo"
	"github.com/dirk/quickhook/tracing"
)

//go:embed pre_commit_git_shim.sh
var PRE_COMMIT_GIT_SHIM string

const PRE_COMMIT_HOOK = "pre-commit"
const PRE_COMMIT_MUTATING_HOOK = "pre-commit-mutating"

const FAILED_EXIT_CODE = 65         // EX_DATAERR - hooks didn't pass
const NOTHING_STAGED_EXIT_CODE = 66 // EX_NOINPUT

type PreCommit struct {
	Repo *repo.Repo
}

// argsFiles can be non-empty with the files passed in by the user when manually running this hook,
// or it can be empty and the list of files will be retrieved from Git.
func (hook *PreCommit) Run(argsFiles []string) error {
	// The shimming is really fast, so just do it first with a defer for cleaning up the
	// temporary directory.
	dirForPath, err := shimGit()
	if err != nil {
		return err
	}
	defer os.RemoveAll(dirForPath)

	files, mutatingExecutables, parallelExecutables, err := internal.FanOut3(
		func() ([]string, error) {
			if len(argsFiles) > 0 {
				return argsFiles, nil
			}
			if files, err := hook.Repo.FilesToBeCommitted(); err != nil {
				return nil, err
			} else {
				return files, nil
			}
		},
		func() ([]string, error) {
			return hook.Repo.FindHookExecutables(PRE_COMMIT_MUTATING_HOOK)
		},
		func() ([]string, error) {
			return hook.Repo.FindHookExecutables(PRE_COMMIT_HOOK)
		},
	)
	if err != nil {
		return err
	}

	stdin := strings.Join(files, "\n")

	// Run mutating executables sequentially.
	for _, executable := range mutatingExecutables {
		result := runExecutable(hook.Repo.Root, executable, os.Environ(), stdin)
		if hook.checkResult(result) {
			os.Exit(FAILED_EXIT_CODE)
		}
	}
	// And the rest in parallel.
	results := lop.Map(parallelExecutables, func(executable string, _ int) hookResult {
		// Insert the git shim's directory into the PATH to prevent usage of git.
		env := append(os.Environ(), fmt.Sprintf("PATH=%s:%s", dirForPath, os.Getenv("PATH")))
		return runExecutable(hook.Repo.Root, executable, env, stdin)
	})
	errored := false
	for _, result := range results {
		errored = hook.checkResult(result) || errored
	}
	if errored {
		os.Exit(FAILED_EXIT_CODE)
	}
	return nil
}

// Returns true if the hook errored, false if it did not.
func (hook *PreCommit) checkResult(result hookResult) bool {
	if result.err == nil {
		// Print any stderr even if the hook executable succeeded.
		result.printStderr()
		return false
	}
	// Maybe print a header?
	result.printStderr()
	result.printStdout()
	return true
}

func shimGit() (string, error) {
	actualGit, err := exec.LookPath("git")
	if err != nil {
		return "", err
	}
	// Trusting that we didn't get a malicious path back from LookPath().
	templated := strings.Replace(PRE_COMMIT_GIT_SHIM, "ACTUAL_GIT", actualGit, 1)

	span := tracing.NewSpan("shim-git")
	defer span.End()

	dir, err := os.MkdirTemp("", "quickhook-git-*")
	if err != nil {
		return "", err
	}

	git := path.Join(dir, "git")
	err = os.WriteFile(git, []byte(templated), 0755)
	if err != nil {
		return "", err
	}
	return dir, nil
}
