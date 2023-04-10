package hooks

import (
	"bytes"
	"io"
	"testing"

	"github.com/creack/pty"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dirk/quickhook/testutils"
)

func initGitForPreCommit(t *testing.T) testutils.TempDir {
	tempDir := testutils.NewTempDir(t, 1)
	tempDir.RequireExec("git", "init", "--quiet", ".")
	tempDir.RequireExec("git", "config", "--local", "user.name", "example")
	tempDir.RequireExec("git", "config", "--local", "user.email", "example@example.com")
	tempDir.WriteFile([]string{"example.txt"}, "Changed!")
	tempDir.RequireExec("git", "add", "example.txt")
	return tempDir
}

func TestFailingHookWithoutPty(t *testing.T) {
	tempDir := initGitForPreCommit(t)
	tempDir.MkdirAll(".quickhook", "pre-commit")
	tempDir.WriteFile(
		[]string{".quickhook", "pre-commit", "fails"},
		"#!/bin/bash \n printf \"first line\\nsecond line\\n\" \n exit 1")

	output, err := tempDir.ExecQuickhook("hook", "pre-commit")
	assert.Error(t, err)
	assert.Equal(t, "fails: first line\nfails: second line\n", output)
}

func TestFailingHookWithPty(t *testing.T) {
	ptyTests := []struct {
		name string
		arg  []string
		out  string
	}{
		{
			"no args",
			[]string{},
			"\x1b[31mfails\x1b[0m: first line\r\n\x1b[31mfails\x1b[0m: second line\r\n",
		},
		{
			"no-color arg",
			[]string{"--no-color"},
			"fails: first line\r\nfails: second line\r\n",
		},
	}
	for _, tt := range ptyTests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := initGitForPreCommit(t)
			tempDir.MkdirAll(".quickhook", "pre-commit")
			tempDir.WriteFile(
				[]string{".quickhook", "pre-commit", "fails"},
				"#!/bin/bash \n printf \"first line\\nsecond line\\n\" \n exit 1",
			)

			cmd := tempDir.NewCommand(
				tempDir.Quickhook,
				append([]string{"hook", "pre-commit"}, tt.arg...)...,
			)
			f, err := pty.Start(cmd)
			require.NoError(t, err)
			defer func() { _ = f.Close() }()

			var b bytes.Buffer
			io.Copy(&b, f)

			assert.Equal(t, tt.out, b.String())
		})
	}
}

func TestPassesWithNoHooks(t *testing.T) {
	tempDir := initGitForPreCommit(t)
	tempDir.MkdirAll(".quickhook", "pre-commit")

	output, err := tempDir.ExecQuickhook("hook", "pre-commit")
	assert.NoError(t, err)
	assert.Equal(t, "", output)
}

func TestPassesWithPassingHooks(t *testing.T) {
	tempDir := initGitForPreCommit(t)
	tempDir.MkdirAll(".quickhook", "pre-commit")
	tempDir.WriteFile(
		[]string{".quickhook", "pre-commit", "passes1"},
		"#!/bin/bash \n echo \"passed\" \n exit 0")
	tempDir.WriteFile(
		[]string{".quickhook", "pre-commit", "passes2"},
		"#!/bin/sh \n echo \"passed\"")

	output, err := tempDir.ExecQuickhook("hook", "pre-commit")
	assert.NoError(t, err)
	assert.Equal(t, "", output)
}

func TestPassesWithNoFilesToBeCommitted(t *testing.T) {
	tempDir := initGitForPreCommit(t)
	tempDir.MkdirAll(".quickhook", "pre-commit")
	tempDir.WriteFile([]string{".quickhook", "pre-commit", "passes"}, "#!/bin/sh \n echo \"passed\"")
	tempDir.RequireExec("git", "commit", "--message", "Commit example.txt", "--quiet", "--no-verify")

	output, err := tempDir.ExecQuickhook("hook", "pre-commit")
	assert.NoError(t, err)
	assert.Equal(t, "", output)
}

func TestHandlesDeletedFiles(t *testing.T) {
	tempDir := initGitForPreCommit(t)
	tempDir.MkdirAll(".quickhook", "pre-commit")
	tempDir.WriteFile([]string{".quickhook", "pre-commit", "passes"}, "#!/bin/sh \n echo \"passed\"")
	tempDir.RequireExec("git", "commit", "--message", "Commit example.txt", "--quiet", "--no-verify")
	tempDir.RequireExec("git", "rm", "example.txt", "--quiet")
	tempDir.WriteFile(
		[]string{"other-example.txt"},
		"Also changed!")
	tempDir.RequireExec("git", "add", "other-example.txt")

	output, err := tempDir.ExecQuickhook("hook", "pre-commit")
	assert.NoError(t, err)
	assert.Equal(t, "", output)
}

func TestShimsGitToDenyAccess(t *testing.T) {
	tempDir := initGitForPreCommit(t)
	tempDir.MkdirAll(".quickhook", "pre-commit")
	tempDir.WriteFile([]string{".quickhook", "pre-commit", "accesses-git"}, "#!/bin/sh \n git status")

	output, err := tempDir.ExecQuickhook("hook", "pre-commit")
	assert.Error(t, err)
	assert.Equal(t, "accesses-git: git is not allowed in parallel hooks (git status)\n", output)
}

func TestMutatingCanAccessGit(t *testing.T) {
	ptyTests := []struct {
		name string
		hook string
		out  string
	}{
		{
			"stdout",
			"#!/bin/sh \n git status",
			"",
		},
		{
			"stderr",
			"#!/bin/sh \n git status 1>&2",
			"accesses-git: On branch master",
		},
	}
	for _, tt := range ptyTests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := initGitForPreCommit(t)
			tempDir.MkdirAll(".quickhook", "pre-commit-mutating")
			tempDir.WriteFile([]string{".quickhook", "pre-commit-mutating", "accesses-git"}, tt.hook)

			output, err := tempDir.ExecQuickhook("hook", "pre-commit")
			assert.NoError(t, err)
			if tt.out == "" {
				assert.Empty(t, output)
			} else {
				assert.Contains(t, output, tt.out)
			}
		})
	}
}
