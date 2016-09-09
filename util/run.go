package util

import (
    "os/exec"
    "strings"

    "github.com/dirk/quickhook/context"
)

type ExecutableResult struct {
	Executable *context.Executable
	CommandError error
	CombinedOutput string
}

func RunExecutable(executable *context.Executable, files []string) *ExecutableResult {
	cmd := exec.Command(executable.AbsolutePath)
	cmd.Stdin = strings.NewReader(strings.Join(files, "\n"))

	combinedOutputBytes, exitError := cmd.CombinedOutput()

	return &ExecutableResult{
		Executable: executable,
		CommandError: exitError,
		CombinedOutput: string(combinedOutputBytes),
	}
}
