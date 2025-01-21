package lexer

import (
	"fmt"
	"reflect"

	"csv/gameconfig/infra/ctype"
)

const (
	Def_PK      = "pk"
	Def_Int32   = "int"
	Def_Int64   = "long"
	Def_Uint32  = "uint"
	Def_Uint64  = "ulong"
	Def_Float32 = "float"
	Def_Float64 = "double"
	Def_String  = "string"
	Def_Bool    = "bool"
	Def_Enum    = "enum"
	Def_Any     = "any"
)

const (
	Sign_Ref        = '@'
	Sign_ArrayStart = '['
	Sign_ArrayEnd   = ']'
)

const (
	SingleLineStart = "{"
	SingleLineEnd   = "}"
)

var simpleTypeMapping = map[string]*GoType{
	Def_PK:      {def: Def_PK, typ: reflect.TypeOf(ctype.ID(0))},
	Def_Int32:   {def: Def_Int32, typ: reflect.TypeOf(int32(0))},
	Def_Int64:   {def: Def_Int64, typ: reflect.TypeOf(int64(0))},
	Def_Uint32:  {def: Def_Uint32, typ: reflect.TypeOf(uint32(0))},
	Def_Uint64:  {def: Def_Uint64, typ: reflect.TypeOf(uint64(0))},
	Def_Float32: {def: Def_Float32, typ: reflect.TypeOf(float32(0))},
	Def_Float64: {def: Def_Float64, typ: reflect.TypeOf(float64(0))},
	Def_String:  {def: Def_String, typ: reflect.TypeOf("")},
	Def_Bool:    {def: Def_Bool, typ: reflect.TypeOf(false)},
	Def_Enum:    {def: Def_Enum, typ: reflect.TypeOf(ctype.ENUM(0))},
	Def_Any:     {def: Def_Any},
}

const (
	RowHeader_Desc = iota
	RowHeader_Name
	RowHeader_Type
	RowHeader_Data
	RowHeader_End_
)

type GoType struct {
	def string
	typ reflect.Type
	ref string
	sub []*GoType
	arr bool
}

type SingleLine struct {
	name  string
	typ   *GoType
	value []string
}

type Column struct {
	name string
	typ  *GoType
}

type Table struct {
	sl   []*SingleLine
	cs   []*Column
	data map[string][]string
	md5  string
}

func InitTable(name string, rows [][]string) (tb *Table, err error) {
	var rowIdx, currRowIdx, columnIdx int
	defer func() {
		tb = nil
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

		err = fmt.Errorf("%s:(%d,%d) %s, %s",
			name, currRowIdx, columnIdx, rows[currRowIdx][columnIdx], err.Error())
	}()

	nRow := len(rows)
	if nRow < RowHeader_End_ {
		panic(fmt.Errorf("invalid row number %d", nRow))
	}

	single := make([]*SingleLine, 0, 2)
	for rowIdx < nRow {
		row := rows[rowIdx]
		if row[0] != SingleLineStart {
			break
		}

		var sl SingleLine
		parseSingleLine(row, &sl)
		single = append(single, &sl)
		rowIdx++
	}

	if nRow-rowIdx < RowHeader_End_ {
		panic(fmt.Errorf("invalid row number %d", nRow))
	}

	nameRow := rows[rowIdx+RowHeader_Name]
	typeRow := rows[rowIdx+RowHeader_Type]
	nColumn := len(nameRow)
	tb = &Table{
		sl: single,
		cs: make([]*Column, 0, nColumn),
	}

	for columnIdx = 0; columnIdx < nColumn; columnIdx++ {
		currRowIdx = rowIdx + RowHeader_Name
		cn := nameRow[columnIdx]
		if cn == "" {
			panic(fmt.Errorf("empty column name"))
		}

		if cn[0] == Sign_ArrayStart {
			if len(cn) < 2 {
				panic(fmt.Errorf("invalid array name %s", cn))
			}

			e := len(cn) - 1
			if cn[e] == Sign_ArrayEnd {
				cn = cn[1:e]
				if cn == "" {
					panic(fmt.Errorf("empty array name"))
				}
			} else {
				cn = cn[1:]
				for columnIdx++; columnIdx < nColumn; columnIdx++ {
					sn := nameRow[columnIdx]
					ln := len(sn)
					if ln == 0 {
						panic(fmt.Errorf("empty array member name"))
					}

					if sn[ln-1] == Sign_ArrayEnd {
						sn = sn[:ln-1]
						if sn == "" {
							panic(fmt.Errorf("empty array member name"))
						}

					}
				}
			}

		}

		currRowIdx = rowIdx + RowHeader_Type
		tb.cs = append(tb.cs, &Column{
			name: cn,
			typ:  toGoType(typeRow[columnIdx]),
		})

	}

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

func toGoType(typ string) *GoType {
	gt, ok := simpleTypeMapping[typ]
	if !ok {
		if len(typ) > 2 && typ[0] == Sign_Ref {
			return &GoType{
				def: "",
				typ: reflect.TypeOf(ctype.ID(0)),
				ref: typ[1:],
			}
		}

		panic(fmt.Errorf("invalid type %s", typ))
	}

	return gt
}
