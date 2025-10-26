package service

import (
	"io"
	"os/exec"
)

func Ps(output io.Writer) error {
	cmd := exec.Command("ps", "aux")
	cmdOutput, err := cmd.Output()
	if err != nil {
		return err
	}
	_, err = output.Write(cmdOutput)
	return err
}
