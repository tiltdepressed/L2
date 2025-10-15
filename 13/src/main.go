package main

import (
	"13/cut"
	"13/parser"
	"bufio"
	"fmt"
	"os"
)

func main() {
	opt, err := parser.ParseFlags()
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	sc := bufio.NewScanner(os.Stdin)

	buf := make([]byte, 0, 1024*1024)
	sc.Buffer(buf, 10*1024*1024)
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	for sc.Scan() {
		line := sc.Text()

		out, ok := cut.Cut(line, opt)
		if !ok {
			continue
		}
		if _, err := w.WriteString(out); err != nil {

			fmt.Println(err)
			os.Exit(1)
		}
		if err := w.WriteByte('\n'); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

	}

	if err := sc.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
