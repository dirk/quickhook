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

// OS exit code to use when hooks didn't pass
const FAILED_EXIT_CODE = 1

func PreCommit(c *context.Context) error {
	files, err := c.FilesToBeCommitted()
	if err != nil { return err }

	if len(files) == 0 {
		fmt.Println("No files to be committed!")
		return nil
	}

	executables, err := c.ExecutablesForHook(HOOK)
	if err != nil { return err }

	// for _, file := range files {
	// 	fmt.Printf("file: %v\n", file)
	// }

	results := runExecutablesInParallel(executables, files)
	hasErrors := false

	for _, result := range results {
		if result.err != nil {
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
	err error
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

	combinedOutputBytes, exitErr := cmd.CombinedOutput()

	return &Result{
		executablePath: path,
		err: exitErr,
		combinedOutput: string(combinedOutputBytes),
	}
}
