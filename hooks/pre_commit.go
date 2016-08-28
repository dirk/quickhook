package hooks

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/jeffail/tunny"

	"github.com/dirk/quickhook/context"
)

const HOOK = "pre-commit"

const FAILED_EXIT_CODE         = 65 // EX_DATAERR - hooks didn't pass
const NOTHING_STAGED_EXIT_CODE = 66 // EX_NOINPUT

func PreCommit(c *context.Context) error {
	files, err := c.FilesToBeCommitted()
	if err != nil { return err }

	if len(files) == 0 {
		fmt.Println("No files to be committed!")
		os.Exit(NOTHING_STAGED_EXIT_CODE)
	}

	executables, err := c.ExecutablesForHook(HOOK)
	if err != nil { return err }

	results := runExecutablesInParallel(executables, files)
	hasErrors := false

	for _, result := range results {
		if result.commandError != nil {
			hasErrors = true

			fmt.Printf("%v:\n", result.executablePath)

			output := strings.TrimSpace(result.combinedOutput)
			for _, line := range strings.Split(output, "\n") {
				fmt.Printf("  %v\n", line)
			}
		}
	}

	if hasErrors {
		os.Exit(FAILED_EXIT_CODE)
	}

	return nil
}

type Result struct {
	executablePath string
	commandError error
	combinedOutput string
}

// Uses a pool sized to the number of CPUs to run all the executables. It's
// sized to the CPU count so that we fully utilized the hardwire but don't
// context switch in the OS too much.
func runExecutablesInParallel(executables []string, files[]string) ([]*Result) {
	bufferSize := len(executables)

	in := make(chan string, bufferSize)
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

func runExecutable(path string, files []string) *Result {
	cmd := exec.Command(path)
	cmd.Stdin = strings.NewReader(strings.Join(files, "\n"))

	combinedOutputBytes, exitError := cmd.CombinedOutput()

	return &Result{
		executablePath: path,
		commandError: exitError,
		combinedOutput: string(combinedOutputBytes),
	}
}
