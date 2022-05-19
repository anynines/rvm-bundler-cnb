package bundler

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/avarteqgmbh/rvm-cnb/rvm"
)

// BashCmd represents a bash cmd runner
type RunBashCmd struct{}

// NewBashCmd creates a new bash cmd runner
func NewRunBashCmd() RunBashCmd {
	return RunBashCmd{}
}

// RunBashCmd executes a command in an interactive BASH shell
func (r RunBashCmd) RunBashCmd(command string, WorkingDir string) (string, error) {
	logger := rvm.NewLogEmitter(os.Stdout)
	stdout := ""

	cmd := exec.Command("bash")
	cmd.Dir = WorkingDir
	cmd.Args = append(
		cmd.Args,
		"--login",
		"-c",
		strings.Join(
			[]string{
				"source",
				filepath.Join(os.ExpandEnv("$rvm_path"), "profile.d", "rvm"),
				"&&",
				command,
			},
			" ",
		),
	)
	cmd.Env = os.Environ()

	logger.Process("Executing: %s", strings.Join(cmd.Args, " "))

	stdoutPipe, _ := cmd.StdoutPipe()
	var stderrBuf bytes.Buffer
	cmd.Stderr = io.MultiWriter(&stderrBuf)

	if err := cmd.Start(); err != nil {
		logger.Process("Failed to start command: %s", cmd.String())
		logger.Break()
		return "", err
	}

	stdoutReader := bufio.NewReader(stdoutPipe)
	stdoutLine, err := stdoutReader.ReadString('\n')
	for err == nil {
		logger.Subprocess(stdoutLine)
		stdout = fmt.Sprintf("%s%s%s", stdout, stdoutLine, "\n")
		stdoutLine, err = stdoutReader.ReadString('\n')
	}
	err = cmd.Wait()

	if err != nil {
		logger.Process("Command failed: %s", cmd.String())
		logger.Process("Error status code: %s", err.Error())
		if len(stderrBuf.String()) > 0 {
			logger.Process("Command output on stderr:")
			logger.Subprocess(stderrBuf.String())
		}
		return "", err
	}

	logger.Break()

	return stdout, nil
}
