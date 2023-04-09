package hooks

import (
	"os"

	"github.com/dirk/quickhook/repo"
)

const COMMIT_MSG_HOOK = "commit-msg"

type CommitMsg struct {
	Repo *repo.Repo
}

func (hook *CommitMsg) Run(messageFile string) error {
	executables, err := hook.Repo.FindHookExecutables(COMMIT_MSG_HOOK)
	if err != nil {
		return err
	}
	for _, executable := range executables {
		result := runExecutable(hook.Repo.Root, executable, []string{}, "", messageFile)
		if result.err == nil {
			continue
		}
		result.printStderr()
		result.printStdout()
		os.Exit(FAILED_EXIT_CODE)
	}
	return nil
}
