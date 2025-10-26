package service

import (
	"fmt"
	"io"
	"os"
)

func Pwd(output io.Writer) error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(output, pwd)
	return err
}
