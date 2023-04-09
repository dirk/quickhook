package hooks

import (
	// "fmt"
	"bytes"
	"io"
	"os"
	"os/exec"
	"path"
	"testing"

	"github.com/creack/pty"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type tempDir struct {
	test      *testing.T
	root      string
	quickhook string
}

func newTempDir(t *testing.T) tempDir {
	cwd, err := os.Getwd()
	require.NoError(t, err)
	return tempDir{
		test:      t,
		root:      t.TempDir(),
		quickhook: path.Join(cwd, "..", "quickhook"),
	}
}

func (temp *tempDir) newCommand(name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...)
	cmd.Dir = temp.root
	return cmd
}

func (temp *tempDir) requireExec(name string, arg ...string) {
	cmd := temp.newCommand(name, arg...)
	_, err := cmd.Output()
	require.NoError(temp.test, err, cmd)
}

func (temp *tempDir) execHook(hook string, arg ...string) (string, error) {
	cmd := temp.newCommand(
		temp.quickhook,
		append([]string{"hook", hook}, arg...)...,
	)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func (temp *tempDir) writeFile(relativePath []string, data string) {
	fullPath := path.Join(append([]string{temp.root}, relativePath...)...)
	err := os.WriteFile(fullPath, []byte(data), 0755)
	if err != nil {
		temp.test.Fatal(err)
	}
}

func (temp *tempDir) mkdirAll(relativePath ...string) {
	fullPath := path.Join(append([]string{temp.root}, relativePath...)...)
	err := os.MkdirAll(fullPath, 0755)
	if err != nil {
		temp.test.Fatal(err)
	}
}

func initGit(t *testing.T) tempDir {
	temp := newTempDir(t)
	temp.requireExec("git", "init", "--quiet", ".")
	temp.requireExec("git", "config", "--local", "user.name", "example")
	temp.requireExec("git", "config", "--local", "user.email", "example@example.com")
	temp.writeFile([]string{"example.txt"}, "Changed!")
	temp.requireExec("git", "add", "example.txt")
	return temp
}

func TestFailingHookWithoutPty(t *testing.T) {
	temp := initGit(t)
	temp.mkdirAll(".quickhook", "pre-commit")
	temp.writeFile(
		[]string{".quickhook", "pre-commit", "fails"},
		"#!/bin/bash \n printf \"first line\\nsecond line\\n\" \n exit 1",
	)

	output, err := temp.execHook("pre-commit")
	assert.Error(t, err)
	assert.Equal(t, "fails: first line\nfails: second line\n", output)
}

var ptyTests = []struct {
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

func TestFailingHookWithPty(t *testing.T) {
	for _, tt := range ptyTests {
		t.Run(tt.name, func(t *testing.T) {
			temp := initGit(t)
			temp.mkdirAll(".quickhook", "pre-commit")
			temp.writeFile(
				[]string{".quickhook", "pre-commit", "fails"},
				"#!/bin/bash \n printf \"first line\\nsecond line\\n\" \n exit 1",
			)

			cmd := temp.newCommand(
				temp.quickhook,
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
	temp := initGit(t)
	temp.mkdirAll(".quickhook", "pre-commit")

	output, err := temp.execHook("pre-commit")
	assert.NoError(t, err)
	assert.Equal(t, "", output)
}

func TestPassesWithPassingHooks(t *testing.T) {
	temp := initGit(t)
	temp.mkdirAll(".quickhook", "pre-commit")
	temp.writeFile(
		[]string{".quickhook", "pre-commit", "passes1"},
		"#!/bin/bash \n echo \"passed\" \n exit 0",
	)
	temp.writeFile(
		[]string{".quickhook", "pre-commit", "passes2"},
		"#!/bin/sh \n echo \"passed\"",
	)

	output, err := temp.execHook("pre-commit")
	assert.NoError(t, err)
	assert.Equal(t, "", output)
}

func TestPassesWithNoFilesToBeCommitted(t *testing.T) {
	temp := initGit(t)
	temp.mkdirAll(".quickhook", "pre-commit")
	temp.writeFile([]string{".quickhook", "pre-commit", "passes"}, "#!/bin/sh \n echo \"passed\"")
	temp.requireExec("git", "commit", "--message", "Commit example.txt", "--quiet", "--no-verify")

	output, err := temp.execHook("pre-commit")
	assert.NoError(t, err)
	assert.Equal(t, "", output)
}

func TestHandlesDeletedFiles(t *testing.T) {
	temp := initGit(t)
	temp.mkdirAll(".quickhook", "pre-commit")
	temp.writeFile([]string{".quickhook", "pre-commit", "passes"}, "#!/bin/sh \n echo \"passed\"")
	temp.requireExec("git", "commit", "--message", "Commit example.txt", "--quiet", "--no-verify")
	temp.requireExec("git", "rm", "example.txt", "--quiet")
	temp.writeFile(
		[]string{"other-example.txt"},
		"Also changed!",
	)
	temp.requireExec("git", "add", "other-example.txt")

	output, err := temp.execHook("pre-commit")
	assert.NoError(t, err)
	assert.Equal(t, "", output)
}
