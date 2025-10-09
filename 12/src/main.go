package main

import (
	"12/grep"
	"12/parser"
	"fmt"
	"os"
	"strings"
)

func main() {
	opt, err := parser.ParseFlags()
	if err != nil {
		os.Exit(1)
	}

	lines, count, err := grep.Grep(opt)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(2)
	}

	if opt.OnlyStringsCount {
		fmt.Println(count)
		return
	}

	if len(lines) > 0 {
		fmt.Println(strings.Join(lines, "\n"))
	}
}
