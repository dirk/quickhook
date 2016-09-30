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

const PRE_COMMIT_HOOK = "pre-commit"

const FAILED_EXIT_CODE = 65         // EX_DATAERR - hooks didn't pass
const NOTHING_STAGED_EXIT_CODE = 66 // EX_NOINPUT

type PreCommitOpts struct {
	NoColor bool
	Files   []string
	All     bool
}

func (opts *PreCommitOpts) ListFiles(c *context.Context) ([]string, error) {
	if len(opts.Files) > 0 {
		for _, file := range opts.Files {
			isFile, err := context.IsFile(file)
			if err != nil {
				return nil, err
			}

			if !isFile {
				color.Yellow(fmt.Sprintf("File not found: %v", file))
				os.Exit(NOTHING_STAGED_EXIT_CODE)
			}
		}

		return opts.Files, nil
	} else if opts.All {
		return c.AllFiles()
	} else {
		return c.FilesToBeCommitted()
	}
}

func PreCommit(c *context.Context, opts *PreCommitOpts) error {
	if opts.NoColor {
		color.NoColor = true
	}

	files, err := opts.ListFiles(c)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		color.Yellow("No files to be committed!")
		os.Exit(NOTHING_STAGED_EXIT_CODE)
	}

	executables, err := c.ExecutablesForHook(PRE_COMMIT_HOOK)
	if err != nil {
		return err
	}

	results := runExecutablesInParallel(executables, files)
	hasErrors := false

	for _, result := range results {
		fmt.Printf("%v: %v\n", result.executable.Name, errToStringStatus(result.commandError))

		if result.commandError != nil {
			hasErrors = true

			output := strings.TrimSpace(result.combinedOutput)
			if output != "" {
				color.Red(output)
			}
		}
	}

	if hasErrors {
		os.Exit(FAILED_EXIT_CODE)
	}

	return nil
}

type Result struct {
	executable     *context.Executable
	commandError   error
	combinedOutput string
}

// Uses a pool sized to the number of CPUs to run all the executables. It's
// sized to the CPU count so that we fully utilized the hardwire but don't
// context switch in the OS too much.
func runExecutablesInParallel(executables []*context.Executable, files []string) []*Result {
	bufferSize := len(executables)

	in := make(chan *context.Executable, bufferSize)
	out := make(chan *Result, bufferSize)

	pool, err := tunny.CreatePoolGeneric(runtime.NumCPU()).Open()
	if err != nil {
		panic(err)
	}

	defer pool.Close()

	for _, executable := range executables {
		in <- executable

		go func() {
			_, err := pool.SendWork(func() {
				executable := <-in

				out <- runPreCommitExecutable(executable, files)
			})

			// Something real bad happened
			if err != nil {
				panic(err)
			}
		}()
	}

	var results []*Result
	for i := 0; i < bufferSize; i++ {
		results = append(results, <-out)
	}

	return results
}

func runPreCommitExecutable(executable *context.Executable, files []string) *Result {
	cmd := exec.Command(executable.AbsolutePath)
	cmd.Stdin = strings.NewReader(strings.Join(files, "\n"))

	combinedOutputBytes, exitError := cmd.CombinedOutput()

	return &Result{
		executable:     executable,
		commandError:   exitError,
		combinedOutput: string(combinedOutputBytes),
	}
}
