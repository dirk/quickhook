package hooks

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"

	"github.com/dirk/quickhook/repo"
	"github.com/dirk/quickhook/tracing"
)

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
	// Resolve files to be committed in parallel with shimming git.
	filesChan := lo.Async2(func() ([]string, error) {
		if len(argsFiles) > 0 {
			return argsFiles, nil
		}
		if files, err := hook.Repo.FilesToBeCommitted(); err != nil {
			return nil, err
		} else {
			return files, nil
		}
	})
	shimChan := lo.Async2(shimGit)
	mutatingChan := lo.Async2(func() ([]string, error) {
		return hook.Repo.FindHookExecutables(PRE_COMMIT_MUTATING_HOOK)
	})
	parallelChan := lo.Async2(func() ([]string, error) {
		return hook.Repo.FindHookExecutables(PRE_COMMIT_HOOK)
	})

	dirForPath, err := (<-shimChan).Unpack()
	if err != nil {
		return err
	}
	// Check the shimChan first so that if we did successfully create a directory with a shim we
	// can make sure to clean it up if anything else errored.
	defer os.RemoveAll(dirForPath)
	files, err := (<-filesChan).Unpack()
	if err != nil {
		return err
	}
	mutatingExecutables, err := (<-mutatingChan).Unpack()
	if err != nil {
		return err
	}
	parallelExecutables, err := (<-parallelChan).Unpack()
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
	span := tracing.NewSpan("shim-git")
	defer span.End()

	dir, err := os.MkdirTemp("", "quickhook-git-*")
	if err != nil {
		return "", err
	}

	git := path.Join(dir, "git")
	err = os.WriteFile(git, []byte(strings.Join([]string{
		"#!/bin/sh",
		"echo \"git is not allowed in parallel hooks (git $@)\"",
		"exit 1",
		"",
	}, "\n")), 0755)
	if err != nil {
		return "", err
	}
	return dir, nil
}
