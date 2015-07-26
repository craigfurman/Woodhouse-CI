package runner

import (
	"os/exec"
	"syscall"
)

type UnixCommandRunner struct{}

func (UnixCommandRunner) CombinedOutput(cmd *exec.Cmd) ([]byte, uint32, error) {
	out, err := cmd.CombinedOutput()
	if e, ok := err.(*exec.ExitError); ok {
		status := e.Sys().(syscall.WaitStatus)
		return out, uint32(status.ExitStatus()), nil
	}

	return out, 0, err
}
