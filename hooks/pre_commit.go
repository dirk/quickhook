package hooks

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/dirk/quickhook/repo"
	"github.com/dirk/quickhook/tracing"
)

//go:embed pre_commit_git_shim.sh
var PRE_COMMIT_GIT_SHIM string

const PRE_COMMIT_HOOK = "pre-commit"
const PRE_COMMIT_MUTATING_HOOK = "pre-commit-mutating"

// Following Berkeley error codes: https://github.com/openbsd/src/blob/master/include/sysexits.h
const EX_OK = 0
const EX_DATAERR = 65 // hooks didn't pass
const EX_NOINPUT = 66

type PreCommit struct {
	Repo *repo.Repo
}

// argsFiles can be non-empty with the files passed in by the user when manually running this hook,
// or it can be empty and the list of files will be retrieved from Git.
func (hook *PreCommit) Run(argsFiles []string) (int, error) {
	// The shimming is really fast, so just do it first with a defer for cleaning up the
	// temporary directory.
	dirForPath, err := shimGit()
	if err != nil {
		return EX_OK, err
	}
	defer os.RemoveAll(dirForPath)

	var files, mutatingExecutables, parallelExecutables []string
	eg, _ := errgroup.WithContext(context.Background())
	eg.Go(func() error {
		if len(argsFiles) > 0 {
			files = argsFiles
			return nil
		}
		var err error
		files, err = hook.Repo.FilesToBeCommitted()
		return err
	})
	eg.Go(func() (err error) {
		mutatingExecutables, err = hook.Repo.FindHookExecutables(PRE_COMMIT_MUTATING_HOOK)
		return err
	})
	eg.Go(func() (err error) {
		parallelExecutables, err = hook.Repo.FindHookExecutables(PRE_COMMIT_HOOK)
		return err
	})
	if err := eg.Wait(); err != nil {
		return EX_OK, err
	}

	stdin := strings.Join(files, "\n")

	// Run mutating executables sequentially.
	for _, executable := range mutatingExecutables {
		result := runExecutable(hook.Repo.Root, executable, os.Environ(), stdin)
		if hook.checkResult(result) {
			return EX_DATAERR, nil
		}
	}
	// And the rest in parallel.
	var wg sync.WaitGroup
	results := make([]hookResult, len(parallelExecutables))
	for i, executable := range parallelExecutables {
		wg.Go(func() {
			// Insert the git shim's directory into the PATH to prevent usage of git.
			env := append(os.Environ(), fmt.Sprintf("PATH=%s:%s", dirForPath, os.Getenv("PATH")))
			results[i] = runExecutable(hook.Repo.Root, executable, env, stdin)
		})
	}
	wg.Wait()
	errored := false
	for _, result := range results {
		errored = hook.checkResult(result) || errored
	}
	if errored {
		return EX_DATAERR, nil
	}
	return EX_OK, nil
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
