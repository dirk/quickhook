package hooks

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fatih/color"

	"github.com/dirk/quickhook/tracing"
)

func runExecutable(root, executable string, env []string, stdin string, arg ...string) hookResult {
	dir, command := filepath.Split(executable)
	span := tracing.NewSpan(fmt.Sprintf("hook %s %s", filepath.Base(filepath.Clean(dir)), command))
	defer span.End()
	cmd := hookCommand(filepath.Join(root, executable), arg...)
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

func hookCommand(executable string, arg ...string) *exec.Cmd {
	switch strings.ToLower(filepath.Ext(executable)) {
	case ".bat", ".cmd":
		return exec.Command("cmd", append([]string{"/C", executable}, arg...)...)
	case ".ps1":
		return exec.Command("powershell", append([]string{"-NoProfile", "-ExecutionPolicy", "Bypass", "-File", executable}, arg...)...)
	case ".sh":
		return exec.Command("sh", append([]string{executable}, arg...)...)
	}
	if runtime.GOOS == "windows" && hasShellShebang(executable) {
		return exec.Command("sh", append([]string{executable}, arg...)...)
	}
	return exec.Command(executable, arg...)
}

// Potential caveat: script with "bash" in shebang will be
// still executed with "sh". If it rely on Bash-only syntax,
// it will fail. Same with Python and other interpreters.
// This behaviour is Windows-only.
func hasShellShebang(executable string) bool {
	file, err := os.Open(executable)
	if err != nil {
		return false
	}
	defer file.Close()

	buffer := make([]byte, 128)
	n, err := file.Read(buffer)
	if err != nil && n == 0 {
		return false
	}
	firstLine := strings.SplitN(string(buffer[:n]), "\n", 2)[0]
	return strings.HasPrefix(firstLine, "#!") && strings.Contains(firstLine, "sh")
}

type hookResult struct {
	executable string
	stdout     string
	stderr     string
	err        error
}

func (result *hookResult) printStdout() {
	prefix := color.RedString("%s", filepath.Base(result.executable))
	result.printLines(prefix, result.stdout)
}

func (result *hookResult) printStderr() {
	prefix := color.YellowString("%s", filepath.Base(result.executable))
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
