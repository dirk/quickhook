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
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(outputBytes), "\n")

	return filterLinesForFiles(lines)
}

func (c *Context) AllFiles() ([]string, error) {
	cmd := exec.Command("git", "ls-files")

	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

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
		if len(file) == 0 {
			continue
		}

		isFile, err := IsFile(file)
		if err != nil {
			return nil, err
		}

		if isFile {
			files = append(files, file)
		}
	}

	return files, nil
}

type Executable struct {
	Name         string
	RelativePath string
	AbsolutePath string
}

func (c *Context) ExecutablesForHook(hook string) ([]*Executable, error) {
	shortPath := path.Join(".quickhook", hook)
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
		if fileInfo.IsDir() {
			continue
		}

		name := fileInfo.Name()

		if (fileInfo.Mode() & 0111) > 0 {
			relativePath := path.Join(shortPath, name)

			executables = append(executables, &Executable{
				Name:         name,
				RelativePath: relativePath,
				AbsolutePath: path.Join(c.path, relativePath),
			})
		} else {
			fmt.Printf("Warning: Non-executable file found in %v: %v\n", shortPath, name)
		}
	}

	return executables, nil
}

func (c *Context) ListHooks() ([]string, error) {
	hooksPath := path.Join(c.path, ".quickhook")

	entries, err := ioutil.ReadDir(hooksPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Missing hooks directory: %v\n", hooksPath)
			os.Exit(66) // EX_NOINPUT
		} else {
			return nil, err
		}
	}

	var hooks []string
	for _, entry := range entries {
		if entry.IsDir() && isHook(entry.Name()) {
			hooks = append(hooks, entry.Name())
		}
	}

	return hooks, nil
}

func isHook(name string) bool {
	switch name {
	case
		"pre-commit",
		"commit-msg":
		return true
	}

	return false
}

func shimCommandForHook(hook string) (string, error) {
	var args string

	switch hook {
	case "pre-commit":
		args = "pre-commit"
	case "commit-msg":
		args = "commit-msg $1"
	default:
		return "", fmt.Errorf("invalid hook: %v", hook)
	}

	return fmt.Sprintf("quickhook hook %v", args), nil
}

func PathForShim(hook string) string {
	return path.Join(".git", "hooks", hook)
}

func (c *Context) InstallShim(hook string, prompt bool) error {
	shimPath := PathForShim(hook)

	command, err := shimCommandForHook(hook)
	if err != nil {
		return err
	}

	lines := []string{
		"#!/bin/sh",
		command,
		"", // So we get a trailing newline when we join
	}

	file, err := os.Create(shimPath)
	if err != nil {
		return err
	}
	defer file.Close()

	err = os.Chmod(shimPath, 0755)
	if err != nil {
		return err
	}

	file.WriteString(strings.Join(lines, "\n"))

	return nil
}
