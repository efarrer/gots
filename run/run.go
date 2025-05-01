package run

import (
	"bytes"
	"fmt"
	"os/exec"
)

// RunWithOutput runs a commands and returns stdout, and stderr and any error if it failed.
func RunWithOutput(name string, arg ...string) (string, string, error) {
	cmd := exec.Command(name, arg...)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	err := cmd.Start()
	if err != nil {
		return stdoutBuf.String(), stderrBuf.String(), fmt.Errorf("start failed: %w", err)
	}

	err = cmd.Wait()
	if err != nil {
		return stdoutBuf.String(), stderrBuf.String(), fmt.Errorf("command failed: %w", err)
	}

	return stdoutBuf.String(), stderrBuf.String(), nil
}
