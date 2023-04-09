package repo

import (
	"os"
	"path"
	"strings"

	"github.com/samber/lo"
)

func (repo *Repo) FilesToBeCommitted() ([]string, error) {
	lines, err := repo.ExecCommandLines("git", "diff", "--name-only", "--cached")
	if err != nil {
		return nil, err
	}
	return lo.Filter(lines, func(line string, index int) bool {
		isFile, _ := repo.isFile(line)
		return isFile
	}), err
}

func (repo *Repo) ShimGit() (string, error) {
	dir, err := os.MkdirTemp("", "quickhook-git-*")
	if err != nil {
		return "", err
	}

	git := path.Join(dir, "git")
	err = os.WriteFile(git, []byte(strings.Join([]string{
		"#!/bin/sh",
		"echo \"git is not allowed in parallel hooks (git $@)\"",
		"exit 1",
		"",
	}, "\n")), 0755)
	if err != nil {
		return "", err
	}

	return dir, nil
}
