// Package sortfuncs
package sortfuncs

import (
	"bufio"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

func ReadFileLines(filePath string) ([][]string, error) {
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
	var lines [][]string

	for scanner.Scan() {
		line := scanner.Text()
		columns := strings.Split(line, "\t")

		lines = append(lines, columns)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func SortByColumn(lines [][]string, k int, r, n bool) {
	sort.Slice(lines, func(i, j int) bool {
		colIndex := k - 1
		if len(lines[i]) <= colIndex || len(lines[j]) <= colIndex {
			return false
		}

		valI := lines[i][colIndex]
		valJ := lines[j][colIndex]

		var less bool

		if n {
			numI, errI := strconv.Atoi(valI)
			numJ, errJ := strconv.Atoi(valJ)

			if errI == nil && errJ == nil {
				less = numI < numJ
			} else {
				less = valI < valJ
			}
		} else {
			less = valI < valJ
		}

		if r {
			return !less
		}
		return less
	})
}

func SortDefault(lines [][]string, r bool) {
	sort.Slice(lines, func(i, j int) bool {
		lineI := strings.Join(lines[i], "\t")
		lineJ := strings.Join(lines[j], "\t")
		if r {
			return lineI > lineJ
		}
		return lineI < lineJ
	})
}

func GetUniqueLines(lines [][]string) [][]string {
	seenWord := make(map[string]struct{})
	var result [][]string

	for _, line := range lines {
		key := strings.Join(line, "\x00")
		if _, ok := seenWord[key]; !ok {
			seenWord[key] = struct{}{}
			result = append(result, line)
		}
	}
	return result
}
