package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	RowHeader_Desc = iota
	RowHeader_Name
	RowHeader_Type
	RowHeader_End_
)

var typeMapping = map[string]string{
	"int":    "int32",
	"long":   "int64",
	"float":  "float32",
	"double": "float64",
	"string": "string",
	"bool":   "bool",
	"any":    "any",
}

const (
	Type_PK = "pk"
)

type Column struct {
	Name   string
	GoType string
}

type Table struct {
	cs   []Column
	Data [][]string
	MD5  string
}

type Source struct {
	tbs map[string]*Table
}

func main() {
	var cfgName string
	flag.StringVar(&cfgName, "c", "test/config/config.json", "")
	f, err := os.Open(cfgName)
	if err != nil {
		panic(err)
	}

	b, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}

	var cfg Config
	err = json.Unmarshal(b, &cfg)
	if err != nil {
		panic(err)
	}
	_ = f.Close()

	var srcPath string
	flag.StringVar(&srcPath, "p", "test", "")
	f, err = os.Open(srcPath)
	if err != nil {
		panic(err)
	}

	files, err := f.Readdir(-1)
	if err != nil {
		panic(err)
	}

	var src Source
	src.tbs = make(map[string]*Table, len(files))

	for _, fi := range files {
		if fi.IsDir() || filepath.Ext(fi.Name()) != ".csv" {
			continue
		}

		file, err := os.Open(filepath.Join(srcPath, fi.Name()))
		reader := csv.NewReader(file)
		rows, err := reader.ReadAll()
		if err != nil {
			panic(err)
		}

		if len(rows) < RowHeader_End_ {
			panic(fmt.Errorf("invalid row data"))
		}

		row := rows[RowHeader_Desc]
		if len(row) == 0 {
			panic(fmt.Errorf("invalid row header"))
		}

		name := strings.TrimSuffix(fi.Name(), filepath.Ext(fi.Name()))
		src.tbs[name] = &Table{
			cs: make([]Column, len(rows)),
		}

		fmt.Println(rows)
	}

	_ = f.Close()
}
