package main

import (
	"encoding/csv"
	"fmt"
	"os"
)

type ColumnDef struct {
	Name   string
	GoType string
}

type Table struct {
	Def          []ColumnDef
	Data         [][]string
	groupPattern [][]string
	kvMerger     map[string]IMerge
	paramsMerger map[string]IMerge
}

func main() {
	file, err := os.Open("")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	fmt.Println(rows)
}
