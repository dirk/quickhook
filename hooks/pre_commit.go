package hooks

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/dirk/quickhook/context"
)

const HOOK = "pre-commit"

func PreCommit(c *context.Context) error {
	files, err := c.FilesToBeCommited()
	if err != nil { return err }

	executables, err := c.ExecutablesForHook(HOOK)
	if err != nil { return err }

	// for _, file := range files {
	// 	fmt.Printf("file: %v\n", file)
	// }

	for _, executable := range executables {
		// fmt.Printf("executable: %v\n", executable)
		result, err := runExecutable(c, executable, files)

		if err != nil {
			fmt.Printf("Error running hook executable %v: %v\n", executable, err)
		} else if result.err != nil {
			output := strings.TrimSpace(result.combinedOutput)

			fmt.Printf("Error: %v\n", result.err)
			fmt.Printf("Output: %v\n", output)
		}

	}

	return nil
}

type Result struct {
	err error
	combinedOutput string
}

func runExecutable(c *context.Context, path string, files []string) (*Result, error) {
	cmd := exec.Command(path)
	cmd.Stdin = strings.NewReader(strings.Join(files, "\n"))

	combinedOutputBytes, exitErr := cmd.CombinedOutput()

	return &Result{
		err: exitErr,
		combinedOutput: string(combinedOutputBytes),
	}, nil
}
