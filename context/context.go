package context

import (
	"github.com/libgit2/git2go"
)

type Context struct {
	repo *git.Repository
}

func NewContext(path string) (*Context, error) {
	repo, err := git.OpenRepository(path)
	if err != nil {
		return nil, err
	}

	context := &Context{
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
