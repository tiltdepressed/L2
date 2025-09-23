// Package parser
package parser // Этот файл теперь в пакете 'parser'

import (
	"fmt"
	"os"

	"github.com/akamensky/argparse"
)

// SortOptions теперь живет здесь, в пакете parser.
// Структура должна начинаться с большой буквы, чтобы быть видимой в main.
type SortOptions struct {
	Column               int
	NumericSort          bool
	Reverse              bool
	Unique               bool
	MonthSort            bool
	IgnoreTrailingBlanks bool
	CheckSorted          bool
	HumanNumericSort     bool
	FilePath             string
}

// ParseFlags также должна быть с большой буквы, чтобы быть видимой.
func ParseFlags() (*SortOptions, error) {
	parser := argparse.NewParser("sort", "Утилита для сортировки строк в файле")

	k := parser.Int("k", "key", &argparse.Options{Help: "Сортировать по столбцу N", Default: 0})
	n := parser.Flag("n", "numeric-sort", &argparse.Options{Help: "Сортировать по числовому значению"})
	r := parser.Flag("r", "reverse", &argparse.Options{Help: "Сортировать в обратном порядке"})
	u := parser.Flag("u", "unique", &argparse.Options{Help: "Выводить только уникальные строки"})
	M := parser.Flag("M", "month-sort", &argparse.Options{Help: "Сортировать по названию месяца"})
	b := parser.Flag("b", "ignore-leading-blanks", &argparse.Options{Help: "Игнорировать хвостовые пробелы"})
	c := parser.Flag("c", "check", &argparse.Options{Help: "Проверить, отсортированы ли данные"})
	h := parser.Flag("", "human-numeric-sort", &argparse.Options{Help: "Сортировать числа с учетом суффиксов (K, M, G)"})
	filePath := parser.StringPositional(&argparse.Options{Help: "Путь к файлу для сортировки", Default: ""})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		return nil, err
	}

	options := &SortOptions{
		Column:               *k,
		NumericSort:          *n,
		Reverse:              *r,
		Unique:               *u,
		MonthSort:            *M,
		IgnoreTrailingBlanks: *b,
		CheckSorted:          *c,
		HumanNumericSort:     *h,
		FilePath:             *filePath,
	}

	return options, nil
}
