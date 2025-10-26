package service

import (
	"fmt"
	"io"
	"strings"
)

func Echo(args []string, output io.Writer) error {
	_, err := fmt.Fprintln(output, strings.Join(args, " "))
	return err
}
