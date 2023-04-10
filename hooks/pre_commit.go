package hooks

import (
	"fmt"
	"os"
	"strings"

	lop "github.com/samber/lo/parallel"

	"github.com/dirk/quickhook/repo"
)

const PRE_COMMIT_HOOK = "pre-commit"
const PRE_COMMIT_MUTATING_HOOK = "pre-commit-mutating"

const FAILED_EXIT_CODE = 65         // EX_DATAERR - hooks didn't pass
const NOTHING_STAGED_EXIT_CODE = 66 // EX_NOINPUT

type Opts struct {
	NoColor bool
}

type PreCommit struct {
	Repo *repo.Repo
	Opts
}

func (hook *PreCommit) Run(files []string) error {
	dirForPath, err := hook.Repo.ShimGit()
	if err != nil {
		return err
	}
	defer os.RemoveAll(dirForPath)

	stdin := strings.Join(files, "\n")

	// Find any mutating hooks and run them first sequentially.
	mutatingExecutables, err := hook.Repo.FindHookExecutables(PRE_COMMIT_MUTATING_HOOK)
	if err != nil {
		return err
	}
	for _, executable := range mutatingExecutables {
		result := runExecutable(hook.Repo.Root, executable, os.Environ(), stdin)
		if hook.checkResult(result) {
			os.Exit(FAILED_EXIT_CODE)
		}
	}

	parallelExecutables, err := hook.Repo.FindHookExecutables(PRE_COMMIT_HOOK)
	if err != nil {
		return err
	}
	// Run hook executables in parallel.
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
