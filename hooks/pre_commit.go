package hooks

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/jeffail/tunny"

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

	results := runExecutablesInParallel(executables, files)

	for _, result := range results {
		if result.err != nil {
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
		err: exitErr,
		combinedOutput: string(combinedOutputBytes),
	}
}
