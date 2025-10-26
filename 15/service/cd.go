package service

import (
	"os"
)

func Cd(dir string) error {
	err := os.Chdir(dir)
	if err != nil {
		return err
	}
	return nil
}
