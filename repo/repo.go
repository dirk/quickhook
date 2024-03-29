package repo

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/dirk/quickhook/tracing"
)

type Repo struct {
	// Root directory of the repository.
	Root string
}

func NewRepo() (*Repo, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	return &Repo{
		Root: strings.TrimSpace(string(output)),
	}, nil
}

func (repo *Repo) FindHookExecutables(hook string) ([]string, error) {
	span := tracing.NewSpan("find " + hook)
	defer span.End()

	dir := path.Join(".quickhook", hook)
	var infos []fs.FileInfo
	{
		f, err := os.Open(path.Join(repo.Root, dir))
		if err != nil {
			if os.IsNotExist(err) {
				return []string{}, nil
			}
			return nil, err
		}
		defer f.Close()

		// Using Readdir since it returns FileInfo's that include permissions, whereas ReadDir
		// returns returns DirEntry's which does not.
		infos, err = f.Readdir(-1)
		if err != nil {
			return nil, err
		}
	}

	hooks := []string{}
	for _, info := range infos {
		if info.IsDir() {
			continue
		}
		name := info.Name()
		if (info.Mode() & 0111) != 0 {
			hooks = append(hooks, path.Join(dir, name))
		} else {
			fmt.Fprintf(os.Stderr, "Warning: Non-executable file found in %v: %v\n", dir, name)
		}
	}
	return hooks, nil
}

// Runs a command with the repo root as the current working directory. Returns the command's
// standard output with whitespace trimmed.
func (repo *Repo) ExecCommand(name string, arg ...string) (string, error) {
	cmd := exec.Command(name, arg...)
	cmd.Dir = repo.Root
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// Runs ExecCommand and splits its output on newlines.
func (repo *Repo) ExecCommandLines(name string, arg ...string) ([]string, error) {
	output, err := repo.ExecCommand(name, arg...)
	if err != nil {
		return nil, err
	}
	return strings.Split(output, "\n"), nil
}

func (repo *Repo) isFile(name string) (bool, error) {
	stat, err := os.Stat(path.Join(repo.Root, name))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return !stat.IsDir(), nil
}
