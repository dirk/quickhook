package repo

import (
	"slices"

	"github.com/dirk/quickhook/tracing"
)

func (repo *Repo) FilesToBeCommitted() ([]string, error) {
	span := tracing.NewSpan("git diff")
	defer span.End()
	lines, err := repo.ExecCommandLines("git", "diff", "--name-only", "--cached")
	if err != nil {
		return nil, err
	}
	return slices.DeleteFunc(lines, func(line string) bool {
		isFile, _ := repo.isFile(line)
		return !isFile
	}), nil
}
