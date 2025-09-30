package main

import (
	"fmt"
	"sort"
)

func main() {
	// for idx, val := range "РoМa" {
	// 	fmt.Println(idx, val)
	// }
	line := "zxc"
	fmt.Println(sort.Sort(line))
}
