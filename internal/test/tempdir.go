package test

import (
	"os"
	"os/exec"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

type TempDir struct {
	t         *testing.T
	Root      string
	Quickhook string
}

// Depth should be the depth of the tests from the package root: this is needed
// to correctly find the built quickhook binary for integration testing.
func NewTempDir(t *testing.T, depth int) TempDir {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	elem := []string{cwd}
	for i := 0; i < depth; i++ {
		elem = append(elem, "..")
	}
	elem = append(elem, "quickhook")

	return TempDir{
		t:         t,
		Root:      t.TempDir(),
		Quickhook: path.Join(elem...),
	}
}

func (tempDir *TempDir) NewCommand(name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...)
	cmd.Dir = tempDir.Root
	return cmd
}

func (tempDir *TempDir) RequireExec(name string, arg ...string) {
	cmd := tempDir.NewCommand(name, arg...)
	_, err := cmd.Output()
	require.NoError(tempDir.t, err, cmd)
}

func (tempDir *TempDir) ExecQuickhook(arg ...string) (string, error) {
	cmd := tempDir.NewCommand(tempDir.Quickhook, arg...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func (tempDir *TempDir) WriteFile(relativePath []string, data string) {
	fullPath := path.Join(append([]string{tempDir.Root}, relativePath...)...)
	err := os.WriteFile(fullPath, []byte(data), 0755)
	if err != nil {
		tempDir.t.Fatal(err)
	}
}

func (tempDir *TempDir) MkdirAll(relativePath ...string) {
	fullPath := path.Join(append([]string{tempDir.Root}, relativePath...)...)
	err := os.MkdirAll(fullPath, 0755)
	if err != nil {
		tempDir.t.Fatal(err)
	}
}
