package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"builder/infra"
)

const (
	RowHeader_Desc = iota
	RowHeader_Name
	RowHeader_Type
	RowHeader_Data
	RowHeader_End_
)

const (
	DefType_PK      = "pk"
	DefType_Int32   = "int"
	DefType_Int64   = "long"
	DefType_Float32 = "float"
	DefType_Float64 = "double"
	DefType_String  = "string"
	DefType_Bool    = "bool"
	DefType_Enum    = "enum"
	DefType_Any     = "any"
)

const (
	Sign_Ref        = '@'
	Sign_ArrayStart = '['
	Sign_ArrayEnd   = ']'
)

var goTypeMapping = map[string]GoType{
	DefType_PK:      {def: "pk", typ: reflect.TypeOf(infra.ID(0))},
	DefType_Int32:   {def: "int", typ: reflect.TypeOf(int32(0))},
	DefType_Int64:   {def: "long", typ: reflect.TypeOf(int64(0))},
	DefType_Float32: {def: "float", typ: reflect.TypeOf(float32(0))},
	DefType_Float64: {def: "double", typ: reflect.TypeOf(float64(0))},
	DefType_String:  {def: "string", typ: reflect.TypeOf("")},
	DefType_Bool:    {def: "bool", typ: reflect.TypeOf(bool(false))},
	DefType_Enum:    {def: "enum", typ: reflect.TypeOf(infra.ENUM(0))},
	DefType_Any:     {def: "any"},
}

const (
	SingleLineStart = "{"
	SingleLineEnd   = "}"
)

type GoType struct {
	def string
	typ reflect.Type
	ref string
}

type SingleLine struct {
	name  string
	typ   GoType
	value []string
}

type Column struct {
	name string
	typ  GoType
}

type Table struct {
	sl   []*SingleLine
	cs   []*Column
	data map[string][]string
	md5  string
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

		if len(rows) == 0 {
			panic(fmt.Errorf("empty file %s", fi.Name()))
		}

		name := strings.TrimSuffix(fi.Name(), filepath.Ext(fi.Name()))
		src.tbs[name], err = InitTable(name, rows)
		if err != nil {
			panic(err)
		}

	}

	_ = f.Close()
}

func InitTable(name string, rows [][]string) (tb *Table, err error) {
	var idx int
	defer func() {
		tb = nil
		err = fmt.Errorf("[table %s, row %d] %e", name, idx, recoverError())
	}()

	nRow := len(rows)
	if nRow < RowHeader_End_ {
		return nil, fmt.Errorf("invalid row number %d", nRow)
	}

	single := make([]*SingleLine, 0, 2)
	for idx < nRow {
		row := rows[idx]
		if row[0] != SingleLineStart {
			break
		}

		var sl SingleLine
		parseSingleLine(row, &sl)
		single = append(single, &sl)
		idx++
	}

	if nRow-idx < RowHeader_End_ {
		return nil, fmt.Errorf("invalid row number %d", nRow)
	}

	nameRow := rows[idx+RowHeader_Name]
	typeRow := rows[idx+RowHeader_Type]
	nColumn := len(nameRow)
	tb = &Table{
		sl: single,
		cs: make([]*Column, nColumn),
	}

	for i := 0; i < nColumn; i++ {
		col := tb.cs[i]
		col.name = nameRow[i]
		if col.name == "" {
			return nil, fmt.Errorf("empty name of column %d", i)
		}

		col.typ = toGoType(typeRow[i])
	}

	t := reflect.TypeOf(int32(1))

	return tb, nil
}

// {,name,enum,chest,equip,hero_chip,}
// {,period,int,10,}
func parseSingleLine(row []string, sl *SingleLine) {
	nL := len(row)
	if nL < 5 {
		panic(fmt.Errorf("invalid single line %v", row))
	}

	sl.name = row[1]
	sl.typ = toGoType(row[2])

	var e int
	for i := 4; i < nL; i++ {
		if row[i] == SingleLineEnd {
			e = i
			break
		}
	}

	if e == 0 {
		panic(fmt.Errorf("not found single line end flag"))
	}

	sl.value = row[3:e]
}

func toGoType(typ string) GoType {
	gt, ok := goTypeMapping[typ]
	if !ok {
		if len(typ) > 2 && typ[0] == Sign_Ref {
			return GoType{
				def: "",
				typ: reflect.TypeOf(infra.ID(0)),
				ref: typ[1:],
			}
		}

		panic(fmt.Errorf("invalid type %s", typ))
	}

	return gt
}

func recoverError() (err error) {
	if e := recover(); e != nil {
		err = e.(error)
		if err == nil {
			es := e.(string)
			if es != "" {
				err = fmt.Errorf("%s", es)
			} else {
				err = fmt.Errorf("unknown error found %v", e)
			}
		}
	}

	return
}
