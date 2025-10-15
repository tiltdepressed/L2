// Package parser
package parser

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/akamensky/argparse"
)

type CutOptions struct {
	Fields    map[int]struct{}
	Delimeter string
	Separated bool
}

func ParseFlags() (*CutOptions, error) {
	p := argparse.NewParser("cut", "Аналог cut: вывод указанных полей строк из STDIN")

	S := p.Flag("s", "separated", &argparse.Options{
		Help: "Только строки, содержащие разделитель (строки без разделителя не выводятся).",
	})

	D := p.String("d", "delimeter", &argparse.Options{
		Help:     "Разделитель. По умолчанию — табуляция ('\\t').",
		Required: false,
	})

	F := p.String("f", "fields", &argparse.Options{
		Help:     "Номера полей (через запятую), можно диапазоны. Пример: 1,3-5",
		Required: true,
	})

	if err := p.Parse(os.Args); err != nil {
		fmt.Print(p.Usage(err))
		return nil, err
	}

	if *D == "" {
		*D = "\t"
	}

	fields, err := parseFieldsString(*F)
	if err != nil {
		return nil, err
	}

	return &CutOptions{
		Fields:    fields,
		Delimeter: *D,
		Separated: *S,
	}, nil
}

func parseFieldsString(fieldsString string) (map[int]struct{}, error) {
	tokens := strings.Split(fieldsString, ",")
	result := make(map[int]struct{}, len(tokens))

	for _, tok := range tokens {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			return nil, fmt.Errorf("empty field token in %q", fieldsString)
		}

		dashCount := strings.Count(tok, "-")
		switch dashCount {
		case 0:
			n, err := strconv.Atoi(tok)
			if err != nil {
				return nil, fmt.Errorf("invalid field number %q: %w", tok, err)
			}
			if n <= 0 {
				return nil, fmt.Errorf("field numbers must be >= 1, got %d", n)
			}
			result[n] = struct{}{}
		case 1:
			parts := strings.SplitN(tok, "-", 2)
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])
			if left == "" || right == "" {
				return nil, fmt.Errorf("open ranges not supported: %q", tok)
			}

			a, err1 := strconv.Atoi(left)
			b, err2 := strconv.Atoi(right)
			if err1 != nil || err2 != nil {
				return nil, fmt.Errorf("invalid range %q: %v %v", tok, err1, err2)
			}
			if a <= 0 || b <= 0 {
				return nil, fmt.Errorf("range bounds must be >= 1: %q", tok)
			}
			if a > b {
				return nil, fmt.Errorf("range start must be <= end: %q", tok)
			}

			// ВНИМАНИЕ: большие диапазоны могут быть дорогими по памяти/времени.
			for i := a; i <= b; i++ {
				result[i] = struct{}{}
			}
		default:
			return nil, fmt.Errorf("invalid token with multiple dashes: %q", tok)
		}
	}

	return result, nil
}
