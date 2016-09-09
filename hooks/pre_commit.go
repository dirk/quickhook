package hooks

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/jeffail/tunny"

	"github.com/dirk/quickhook/context"
	"github.com/dirk/quickhook/util"
)

const PRE_COMMIT_HOOK = "pre-commit"

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

	executables, err := c.ExecutablesForHook(PRE_COMMIT_HOOK)
	if err != nil { return err }

	results := runExecutablesInParallel(executables, files)
	hasErrors := false

	for _, result := range results {
		if result.CommandError != nil {
			hasErrors = true

			fmt.Printf("%v:\n", result.Executable.Name)

			output := strings.TrimSpace(result.CombinedOutput)
			color.Red(output)
		}
	}

	if hasErrors {
		os.Exit(FAILED_EXIT_CODE)
	}

	return nil
}

// Uses a pool sized to the number of CPUs to run all the executables. It's
// sized to the CPU count so that we fully utilized the hardwire but don't
// context switch in the OS too much.
func runExecutablesInParallel(executables []*context.Executable, files[]string) ([]*util.ExecutableResult) {
	bufferSize := len(executables)

	in := make(chan *context.Executable, bufferSize)
	out := make(chan *util.ExecutableResult, bufferSize)

	pool, err := tunny.CreatePoolGeneric(runtime.NumCPU()).Open()
	if err != nil { panic(err) }

	defer pool.Close()

	for _, executable := range executables {
		in <- executable

		go func() {
			_, err := pool.SendWork(func() {
				executable := <- in

				out <- util.RunExecutable(executable, files)
			})

			// Something real bad happened
			if err != nil { panic(err) }
		}()
	}

	var results []*util.ExecutableResult
	for i := 0; i < bufferSize; i++ {
		results = append(results, <- out)
	}

	return results
}
