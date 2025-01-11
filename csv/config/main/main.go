package main

import (
	"encoding/csv"
	"fmt"
	"os"
)

const (
	Row_Desc = iota
	Row_Name
	Row_Type
	Row_End_
)

type ColumnDef struct {
	Name   string
	GoType string
}

type Table struct {
	Def          []ColumnDef
	Data         [][]string
	groupPattern [][]string
}

func main() {
	file, err := os.Open("test/shop.csv")
	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	if len(rows) < Row_End_ {
		panic(fmt.Errorf("invalid rows metadata"))
	}

	fmt.Println(rows)
}
