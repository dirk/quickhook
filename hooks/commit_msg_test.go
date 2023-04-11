package hooks

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dirk/quickhook/internal/test"
)

func initGitForCommitMsg(t *testing.T) test.TempDir {
	tempDir := test.NewTempDir(t, 1)
	tempDir.RequireExec("git", "init", "--quiet", ".")
	return tempDir
}

func writeCommitEditMsg(t *testing.T, data string) string {
	name := path.Join(t.TempDir(), "COMMIT_EDITMSG")
	err := os.WriteFile(name, []byte(data), 0644)
	require.NoError(t, err)
	return name
}

func TestHookMutatesCommitMsg(t *testing.T) {
	tempDir := initGitForCommitMsg(t)
	tempDir.MkdirAll(".quickhook", "commit-msg")
	tempDir.WriteFile(
		[]string{".quickhook", "commit-msg", "appends"},
		// -n makes echo not emit a trailing newline.
		"#!/bin/bash \n echo -n \" second\" >> $1")

	editMsgFile := writeCommitEditMsg(t, "First")
	_, err := tempDir.ExecQuickhook("hook", "commit-msg", editMsgFile)
	assert.NoError(t, err)

	newEditMsg, err := os.ReadFile(editMsgFile)
	assert.NoError(t, err)
	assert.Equal(t, "First second", string(newEditMsg))
}

func TestFailingHook(t *testing.T) {
	tempDir := initGitForCommitMsg(t)
	tempDir.MkdirAll(".quickhook", "commit-msg")
	tempDir.WriteFile(
		[]string{".quickhook", "commit-msg", "fails"},
		"#!/bin/bash \n echo \"failed\" \n exit 1")

	output, err := tempDir.ExecQuickhook("hook", "commit-msg", writeCommitEditMsg(t, "Test"))
	assert.Error(t, err)
	assert.Equal(t, "fails: failed\n", output)
}
