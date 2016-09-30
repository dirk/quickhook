package hooks

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"

	"github.com/dirk/quickhook/context"
)

const HOOK = "commit-msg"

type CommitMsgOpts struct {
	NoColor         bool
	MessageTempFile string
}

func CommitMsg(c *context.Context, opts *CommitMsgOpts) error {
	if opts.NoColor {
		color.NoColor = true
	}

	executables, err := c.ExecutablesForHook(HOOK)
	if err != nil {
		return err
	}

	for _, executable := range executables {
		output, err := runCommitMsgExecutable(executable, opts.MessageTempFile)

		fmt.Printf("%v: %v\n", executable.Name, errToStringStatus(err))

		if err != nil {
			output := strings.TrimSpace(output)
			if output != "" {
				color.Red(output)
			}

			os.Exit(FAILED_EXIT_CODE)
		}
	}

	return nil
}

func runCommitMsgExecutable(executable *context.Executable, messageTempFile string) (string, error) {
	cmd := exec.Command(executable.AbsolutePath, messageTempFile)

	combinedOutputBytes, err := cmd.CombinedOutput()
	return string(combinedOutputBytes), err
}
