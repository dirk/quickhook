package main

import (
	"fmt"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dirk/quickhook/testutils"
)

func TestInstallPreCommitYes(t *testing.T) {
	tempDir := testutils.NewTempDir(t, 0)
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
	tempDir := testutils.NewTempDir(t, 0)
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
	tempDir := testutils.NewTempDir(t, 0)
	tempDir.RequireExec("git", "init", "--quiet", ".")

	output, err := tempDir.ExecQuickhook("install", "--yes")
	assert.Error(t, err)
	assert.Contains(t, output, "Missing hooks directory")
}
