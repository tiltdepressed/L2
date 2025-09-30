package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

func findAnagrams(oldLines []string) map[string][]string {
	groups := make(map[string][]string)
	firstBySig := make(map[string]string)
	seenWord := make(map[string]struct{})

	for _, cur := range oldLines {
		line := strings.ToLower(strings.TrimSpace(cur))
		if len(line) == 0 {
			continue
		}

		if _, ok := seenWord[line]; ok {
			continue
		}
		seenWord[line] = struct{}{}

		r := []rune(line)
		sort.Slice(r, func(i, j int) bool { return r[i] < r[j] })
		sig := string(r)

		if _, exists := groups[sig]; !exists {
			firstBySig[sig] = line
		}
		groups[sig] = append(groups[sig], line)
	}

	anagramms := make(map[string][]string)
	for key, val := range groups {
		if len(val) == 1 {
			continue
		}
		sort.Strings(groups[key])
		anagramms[firstBySig[key]] = groups[key]
	}

	return anagramms
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanWords)
	var words []string
	// Ctrl+D to stop
	for scanner.Scan() {
		words = append(words, scanner.Text())
	}
	res := findAnagrams(words)
	keys := make([]string, 0, len(res))
	for k := range res {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("%s: %v\n", k, res[k])
	}
}
