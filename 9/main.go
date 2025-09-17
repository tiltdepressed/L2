package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
)

var ErrInvalidString = errors.New("invalid string format")

func stringUnpack(s string) (string, error) {
	if len(s) == 0 {
		return "", nil
	}

	var result strings.Builder
	runes := []rune(s)

	if unicode.IsDigit(runes[0]) {
		return "", ErrInvalidString
	}

	for i := 0; i < len(runes); i++ {
		char := runes[i]
		charToRepeat := char
		count := 1

		if char == '\\' {
			if i+1 >= len(runes) {
				return "", ErrInvalidString
			}
			charToRepeat = runes[i+1]
			i++
		}

		if i+1 < len(runes) && unicode.IsDigit(runes[i+1]) {
			if unicode.IsDigit(charToRepeat) && char != '\\' {
				return "", ErrInvalidString
			}

			i++
			numStr := string(runes[i])
			if i+1 < len(runes) && unicode.IsDigit(runes[i+1]) {
				return "", ErrInvalidString
			}

			count, _ = strconv.Atoi(numStr)
		}

		result.WriteString(strings.Repeat(string(charToRepeat), count))
	}

	return result.String(), nil
}

func main() {
	fmt.Println("Введите строку:")
	var r string
	fmt.Scanln(&r)
	res, err := stringUnpack(r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Результат:", res)
}
