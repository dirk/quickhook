package context

import (
	"fmt"
	"io/ioutil"
	"path"

	"github.com/libgit2/git2go"
)

type Context struct {
	path string
	repo *git.Repository
}

func NewContext(path string) (*Context, error) {
	repo, err := git.OpenRepository(path)
	if err != nil {
		return nil, err
	}

	context := &Context{
		path: path,
		repo: repo,
	}

	return context, nil
}

func (c *Context) FilesToBeCommited() ([]string, error) {
	statusList, err := c.repo.StatusList(&git.StatusOptions{Show: git.StatusShowIndexOnly})
	if err != nil { return nil, err }

	entryCount, err := statusList.EntryCount()
	if err != nil { return nil, err }

	var files []string
	for i := 0; i < entryCount; i++ {
		entry, _ := statusList.ByIndex(i)
		path := entry.HeadToIndex.NewFile.Path

		files = append(files, path)
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
