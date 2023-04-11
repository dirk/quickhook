package repo

import (
	"github.com/samber/lo"

	"github.com/dirk/quickhook/tracing"
)

func (repo *Repo) FilesToBeCommitted() ([]string, error) {
	span := tracing.NewSpan("git diff")
	defer span.End()
	lines, err := repo.ExecCommandLines("git", "diff", "--name-only", "--cached")
	if err != nil {
		return nil, err
	}
	return lo.Filter(lines, func(line string, index int) bool {
		isFile, _ := repo.isFile(line)
		return isFile
	}), err
}
