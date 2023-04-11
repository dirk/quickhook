package main

import (
	"fmt"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dirk/quickhook/internal/test"
	"github.com/dirk/quickhook/repo"
)

func TestInstallPreCommitYes(t *testing.T) {
	tempDir := test.NewTempDir(t, 0)
	tempDir.RequireExec("git", "init", "--quiet", ".")
	tempDir.MkdirAll(".quickhook", "pre-commit")

	output, err := tempDir.ExecQuickhook("install", "--yes")
	assert.NoError(t, err)
	shimPath := path.Join(".git", "hooks", "pre-commit")
	assert.Equal(t,
		fmt.Sprintf("Installed shim %v", shimPath),
		strings.TrimSpace(output))
	assert.FileExists(t,
		path.Join(tempDir.Root, shimPath))
}

func TestInstallPreCommitMutatingYes(t *testing.T) {
	tempDir := test.NewTempDir(t, 0)
	tempDir.RequireExec("git", "init", "--quiet", ".")
	tempDir.MkdirAll(".quickhook", "pre-commit-mutating")

	output, err := tempDir.ExecQuickhook("install", "--yes")
	assert.NoError(t, err)
	shimPath := path.Join(".git", "hooks", "pre-commit")
	assert.Equal(t,
		fmt.Sprintf("Installed shim %v", shimPath),
		strings.TrimSpace(output))
	assert.FileExists(t,
		path.Join(tempDir.Root, shimPath))
}

func TestInstallNoQuickhookDirectory(t *testing.T) {
	tempDir := test.NewTempDir(t, 0)
	tempDir.RequireExec("git", "init", "--quiet", ".")

	output, err := tempDir.ExecQuickhook("install", "--yes")
	assert.Error(t, err)
	assert.Contains(t, output, "Missing hooks directory")
}

func TestPromptForInstall(t *testing.T) {
	ptyTests := []struct {
		name     string
		stdin    string
		expected bool
	}{
		{
			"yes",
			"yes\n",
			true,
		},
		{
			"short yes",
			"y\n",
			true,
		},
		{
			"no",
			"no\n",
			false,
		},
		{
			"short no",
			"n\n",
			false,
		},
		{
			"no input",
			"",
			false,
		},
	}
	for _, tt := range ptyTests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := test.NewTempDir(t, 0)
			repo := &repo.Repo{Root: tempDir.Root}

			stdin := strings.NewReader(tt.stdin)
			shouldInstall, err := promptForInstallShim(stdin, repo, ".git/hooks/pre-commit")
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, shouldInstall)
		})
	}
}
