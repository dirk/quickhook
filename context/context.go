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

	lines := strings.Split(string(outputBytes), "\n")

	return filterLinesForFiles(lines)
}

func (c *Context) AllFiles() ([]string, error) {
	cmd := exec.Command("git", "ls-files")

	outputBytes, err := cmd.CombinedOutput()
	if err != nil { return nil, err }

	lines := strings.Split(string(outputBytes), "\n")

	return filterLinesForFiles(lines)
}

func IsFile(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	return !stat.IsDir(), nil
}

// Filters an array of lines, returns only lines that are valid paths to
// a file that exists in the filesystem.
func filterLinesForFiles(lines []string) ([]string, error) {
	var files []string

	for _, line := range lines {
		file := strings.TrimSpace(line)
		if len(file) == 0 { continue }

		isFile, err := IsFile(file)
		if err != nil { return nil, err }

		if isFile {
			files = append(files, file)
		}
	}

	return files, nil
}

type Executable struct {
	Name string
	RelativePath string
	AbsolutePath string
}

func (c *Context) ExecutablesForHook(hook string) ([]*Executable, error) {
	shortPath    := path.Join(".quickhook", hook)
	absolutePath := path.Join(c.path, shortPath)

	allFiles, err := ioutil.ReadDir(absolutePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Missing hook directory: %v\n", absolutePath)
			os.Exit(66) // EX_NOINPUT
		} else {
			return nil, err
		}
	}

	var executables []*Executable
	for _, fileInfo := range allFiles {
		if fileInfo.IsDir() { continue }

		name := fileInfo.Name()

		if (fileInfo.Mode() & 0111) > 0 {
			relativePath := path.Join(shortPath, name)

			executables = append(executables, &Executable{
				Name: name,
				RelativePath: relativePath,
				AbsolutePath: path.Join(c.path, relativePath),
			})
		} else {
			fmt.Printf("Warning: Non-executable file found in %v: %v\n", shortPath, name)
		}
	}

	return executables, nil
}
