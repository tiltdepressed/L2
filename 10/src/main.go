package main

import (
	"10/parser"
	sortfuncs "10/sortFuncs"
	"errors"
	"fmt"
	"os"
)

func sort(opt *parser.SortOptions) ([][]string, error) {
	lines, err := sortfuncs.ReadFileLines(opt.FilePath)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения файла: %w", err)
	}

	if opt.Unique {
		lines = sortfuncs.GetUniqueLines(lines)
	}

	if opt.Column == 0 {
		sortfuncs.SortDefault(lines, opt.Reverse)
	} else {
		if opt.Column < 0 {
			return nil, errors.New("неверный номер колонны")
		}
		sortfuncs.SortByColumn(lines, opt.Column, opt.Reverse, opt.NumericSort)
	}

	return lines, nil
}

func main() {
	opt, err := parser.ParseFlags()
	if err != nil {
		os.Exit(1)
	}

	fmt.Println("--- Параметры сортировки (из структуры) ---")
	fmt.Printf("Файл для сортировки: %s\n", opt.FilePath)
	fmt.Printf("Сортировать по колонке: %d\n", opt.Column)
	fmt.Printf("Числовая сортировка (-n): %v\n", opt.NumericSort)
	fmt.Printf("Обратный порядок (-r): %v\n", opt.Reverse)
	fmt.Printf("Только уникальные (-u): %v\n", opt.Unique)

	sortedLines, err := sort(opt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nПрочитано и обработано %d строк.\n", len(sortedLines))
	fmt.Println(sortedLines)
}
