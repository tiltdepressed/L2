package service

import (
	"os"
)

func Kill(pid int) error {
	cmd, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	err = cmd.Kill()
	if err != nil {
		return err
	}
	return nil
}
