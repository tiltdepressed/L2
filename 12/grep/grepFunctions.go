// Package grep
package grep

import (
	"12/parser"
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

func ReadFileLines(filePath string) ([]string, error) {
	var reader io.Reader

	if filePath == "" {
		reader = os.Stdin
	} else {
		file, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		reader = file
	}

	scanner := bufio.NewScanner(reader)
	maxLine := 10 * 1024 * 1024
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, maxLine)
	var lines []string

	for scanner.Scan() {
		line := scanner.Text()

		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func Grep(opt *parser.GrepOptions) ([]string, int, error) {
	lines, err := ReadFileLines(opt.FilePath)
	if err != nil {
		return nil, 0, err
	}

	var matcher func(string) bool
	if opt.FixedString {
		pattern := opt.Pattern
		if opt.IgnoreRegister {
			pattern = strings.ToLower(pattern)
			matcher = func(s string) bool {
				return strings.Contains(strings.ToLower(s), pattern)
			}
		} else {
			matcher = func(s string) bool {
				return strings.Contains(s, pattern)
			}
		}
	} else {
		pattern := opt.Pattern
		if opt.IgnoreRegister {
			pattern = "(?i)" + pattern
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, 0, err
		}
		matcher = re.MatchString
	}

	n := len(lines)
	before := max(0, opt.StringsBefore)
	after := max(0, opt.StringsAfter)

	count := 0
	out := make([]string, 0, n)
	printed := make([]bool, n)

	addLine := func(idx int) {
		if idx < 0 || idx >= n {
			return
		}
		if printed[idx] {
			return
		}
		line := lines[idx]
		if opt.PrintStringNumber {
			out = append(out, fmt.Sprintf("%d:%s", idx+1, line))
		} else {
			out = append(out, line)
		}
		printed[idx] = true
	}

	for i := 0; i < n; i++ {
		m := matcher(lines[i])
		selected := m
		if opt.InvertFilter {
			selected = !m
		}

		if selected {
			count++
			if opt.OnlyStringsCount {
				continue
			}
			for j := max(0, i-before); j < i; j++ {
				addLine(j)
			}
			addLine(i)
			for j := i + 1; j < min(n-1, i+after); j++ {
				addLine(j)
			}
		}
	}
	if opt.OnlyStringsCount {
		return nil, count, nil
	}
	return out, count, nil
}
