package hooks

import (
	"fmt"
	"os"
	"strings"

	lop "github.com/samber/lo/parallel"

	"github.com/dirk/quickhook/repo"
)

const PRE_COMMIT_HOOK = "pre-commit"

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

	executables, err := hook.Repo.FindHookExecutables(PRE_COMMIT_HOOK)
	if err != nil {
		return err
	}

	stdin := strings.Join(files, "\n")
	// Run hook executables in parallel.
	results := lop.Map(executables, func(executable string, _ int) hookResult {
		// Insert the git shim's directory into the PATH to prevent usage of git.
		env := append(os.Environ(), fmt.Sprintf("PATH=%s:%s", dirForPath, os.Getenv("PATH")))
		return runExecutable(hook.Repo.Root, executable, env, stdin)
	})

	errored := false
	for _, result := range results {
		if result.err == nil {
			// Print any stderr even if the hook executable succeeded.
			result.printStderr()
			continue
		}
		// Maybe print a header?
		errored = true
		result.printStderr()
		result.printStdout()
	}
	if errored {
		os.Exit(FAILED_EXIT_CODE)
	}
	return nil
}
