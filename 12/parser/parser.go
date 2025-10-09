// Package parser
package parser

import (
	"errors"
	"fmt"
	"os"

	"github.com/akamensky/argparse"
)

type GrepOptions struct {
	StringsAfter      int
	StringsBefore     int
	StringsAround     int
	OnlyStringsCount  bool
	IgnoreRegister    bool
	InvertFilter      bool
	FixedString       bool
	PrintStringNumber bool
	Pattern           string
	FilePath          string
}

func ParseFlags() (*GrepOptions, error) {
	parser := argparse.NewParser("grep", "Утилита для поиска строк в файле")

	A := parser.Int("A", "after-context", &argparse.Options{Help: "После каждой найденной строки дополнительно вывести N строк после неё (контекст)."})
	B := parser.Int("B", "before-context", &argparse.Options{Help: "Вывести N строк до каждой найденной строки."})
	C := parser.Int("C", "context", &argparse.Options{Help: "Вывести N строк контекста вокруг найденной строки (включает и до, и после; эквивалентно -A N -B N)."})
	c := parser.Flag("c", "count", &argparse.Options{Help: "Выводить только то количество строк, что совпадающих с шаблоном (т.е. вместо самих строк — число)."})
	i := parser.Flag("i", "ignore-case", &argparse.Options{Help: "Игнорировать регистр."})
	v := parser.Flag("v", "invert-match", &argparse.Options{Help: "Инвертировать фильтр: выводить строки, не содержащие шаблон."})
	F := parser.Flag("F", "fixed-strings", &argparse.Options{Help: "Воспринимать шаблон как фиксированную строку, а не регулярное выражение (т.е. выполнять точное совпадение подстроки)."})
	n := parser.Flag("n", "line-number", &argparse.Options{Help: "Выводить номер строки перед каждой найденной строкой."})

	pattern := parser.String("p", "pattern", &argparse.Options{
		Help:     "Шаблон для поиска",
		Required: true,
	})

	filePath := parser.String("f", "file", &argparse.Options{
		Help:     "Путь к файлу для поиска",
		Required: true,
	})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		return nil, err
	}

	if *A < 0 || *B < 0 || *C < 0 {
		return nil, errors.New("значения для -A, -B, -C должны быть неотрицательными")
	}

	options := &GrepOptions{
		StringsAfter:      max(*A, *C),
		StringsBefore:     max(*B, *C),
		StringsAround:     *C,
		OnlyStringsCount:  *c,
		IgnoreRegister:    *i,
		InvertFilter:      *v,
		FixedString:       *F,
		PrintStringNumber: *n,
		Pattern:           *pattern,
		FilePath:          *filePath,
	}

	return options, nil
}
