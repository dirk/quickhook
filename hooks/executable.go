package hooks

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/fatih/color"
)

func runExecutable(root, executable string, env []string, stdin string, arg ...string) hookResult {
	cmd := exec.Command(path.Join(root, executable), arg...)
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdin = strings.NewReader(stdin)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	stdout, err := cmd.Output()
	return hookResult{
		executable: executable,
		stdout:     string(stdout),
		stderr:     stderr.String(),
		err:        err,
	}
}

type hookResult struct {
	executable string
	stdout     string
	stderr     string
	err        error
}

func (result *hookResult) printStdout() {
	prefix := color.RedString("%s", path.Base(result.executable))
	result.printLines(prefix, result.stdout)
}

func (result *hookResult) printStderr() {
	prefix := color.YellowString("%s", path.Base(result.executable))
	result.printLines(prefix, result.stderr)
}

func (result *hookResult) printLines(prefix, lines string) {
	lines = strings.TrimSpace(lines)
	if lines == "" {
		return
	}
	for _, line := range strings.Split(lines, "\n") {
		if line != "" {
			line = " " + line
		}
		fmt.Printf("%s:%s\n", prefix, line)
	}
}
