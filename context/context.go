package context

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
)

type Context struct {
	path string
}

func NewContext(path string) (*Context, error) {
	context := &Context{
		path: path,
	}

	return context, nil
}

func (c *Context) FilesToBeCommitted() ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", "--cached")

	outputBytes, err := cmd.CombinedOutput()
	if err != nil { return nil, err }

	output := string(outputBytes)
	lines := strings.Split(output, "\n")

	var files []string

	// Verify that all the lines are *actually* files
	for _, line := range lines {
		file := strings.TrimSpace(line)
		if len(file) == 0 { continue }

		stat, err := os.Stat(file)
		if err != nil { return nil, err }

		if stat.IsDir() {
			return nil, fmt.Errorf("Unexpected directory in list of staged files: %v", file)
		}

		files = append(files, file)
	}

	return files, nil
}

func (c *Context) ExecutablesForHook(hook string) ([]string, error) {
	shortPath    := path.Join(".quickhook", hook)
	absolutePath := path.Join(c.path, shortPath)

	allFiles, err := ioutil.ReadDir(absolutePath)
	if err != nil { return nil, err }

	var executables []string
	for _, fileInfo := range allFiles {
		if fileInfo.IsDir() { continue }

		name := fileInfo.Name()

		if (fileInfo.Mode() & 0111) > 0 {
			executables = append(executables, path.Join(shortPath, name))
		} else {
			fmt.Printf("Warning: Non-executable file found in %v: %v\n", shortPath, name)
		}
	}

	return executables, nil
}
