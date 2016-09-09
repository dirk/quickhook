package hooks

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/jeffail/tunny"

	"github.com/dirk/quickhook/context"
)

const HOOK = "pre-commit"

const FAILED_EXIT_CODE         = 65 // EX_DATAERR - hooks didn't pass
const NOTHING_STAGED_EXIT_CODE = 66 // EX_NOINPUT

type PreCommitOpts struct  {
	NoColor bool
}

func PreCommit(c *context.Context, opts *PreCommitOpts) error {
	color.NoColor = opts.NoColor

	files, err := c.FilesToBeCommitted()
	if err != nil { return err }

	if len(files) == 0 {
		color.Yellow("No files to be committed!")
		os.Exit(NOTHING_STAGED_EXIT_CODE)
	}

	executables, err := c.ExecutablesForHook(HOOK)
	if err != nil { return err }

	results := runExecutablesInParallel(executables, files)
	hasErrors := false

	for _, result := range results {
		if result.commandError != nil {
			hasErrors = true

			fmt.Printf("%v:\n", result.executable.Name)

			output := strings.TrimSpace(result.combinedOutput)
			color.Red(output)
		}
	}

	if hasErrors {
		os.Exit(FAILED_EXIT_CODE)
	}

	return nil
}

type Result struct {
	executable *context.Executable
	commandError error
	combinedOutput string
}

// Uses a pool sized to the number of CPUs to run all the executables. It's
// sized to the CPU count so that we fully utilized the hardwire but don't
// context switch in the OS too much.
func runExecutablesInParallel(executables []*context.Executable, files[]string) ([]*Result) {
	bufferSize := len(executables)

	in := make(chan *context.Executable, bufferSize)
	out := make(chan *Result, bufferSize)

	pool, err := tunny.CreatePoolGeneric(runtime.NumCPU()).Open()
	if err != nil { panic(err) }

	defer pool.Close()

	for _, executable := range executables {
		in <- executable

		go func() {
			_, err := pool.SendWork(func() {
				executable := <- in

				out <- runExecutable(executable, files)
			})

			// Something real bad happened
			if err != nil { panic(err) }
		}()
	}

	var results []*Result
	for i := 0; i < bufferSize; i++ {
		results = append(results, <- out)
	}

	return results
}

func runExecutable(executable *context.Executable, files []string) *Result {
	cmd := exec.Command(executable.AbsolutePath)
	cmd.Stdin = strings.NewReader(strings.Join(files, "\n"))

	combinedOutputBytes, exitError := cmd.CombinedOutput()

	return &Result{
		executable: executable,
		commandError: exitError,
		combinedOutput: string(combinedOutputBytes),
	}
}
