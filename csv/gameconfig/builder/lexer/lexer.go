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
	Def_Custom  = ""
)

const (
	Sign_Ref        = '@'
	Sign_ArrayStart = '['
	Sign_ArrayEnd   = ']'
	Sign_SubSep     = '|'
)

const (
	SingleLineStart = "{"
	SingleLineEnd   = "}"
)

var simpleTypeMapping = map[string]*GoType{
	Def_PK:      {def: Def_PK, typ: reflect.TypeOf(ctype.GID(0))},
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
	sub []*Column
	an  int
}

type GoValue struct {
	typ *GoType
}

type SingleLine struct {
	name  string
	typ   *GoType
	value []string
}

type Column struct {
	name string
	typ  *GoType
	csi  int
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

			err = fmt.Errorf("%s:(%d,%d) \"%s\", %w",
				name, currRowIdx, columnIdx, rows[currRowIdx][columnIdx], err)
		}
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
	if nColumn < 2 {
		panic(fmt.Errorf("invalid column number %d", nColumn))
	}

	tb = &Table{
		sl: single,
		cs: make([]*Column, 0, nColumn),
	}

	typ0 := toGoType(typeRow[0])
	if typ0.def != Def_PK {
		panic(fmt.Errorf("first column MUST be %s", Def_PK))
	}

	tb.cs = append(tb.cs, &Column{
		name: nameRow[0],
		typ:  typ0,
	})

	for columnIdx = 1; columnIdx < nColumn; columnIdx++ {
		currRowIdx = rowIdx + RowHeader_Type
		tn := typeRow[columnIdx]
		typ := toGoType(tn)

		currRowIdx = rowIdx + RowHeader_Name
		cn := nameRow[columnIdx]
		if cn == "" {
			panic(fmt.Errorf("empty column name"))
		}

		csi := columnIdx
		if cn[0] == Sign_ArrayStart {
			if len(cn) < 2 {
				panic(fmt.Errorf("invalid array name %s", cn))
			}

			ce := len(cn) - 1
			if cn[ce] == Sign_ArrayEnd {
				typ.an = 1
				cn = cn[1:ce]
				if cn == "" {
					panic(fmt.Errorf("empty array name"))
				}

				if typ.def == Def_Custom {
					currRowIdx = rowIdx + RowHeader_Type
					panic(fmt.Errorf("no field of custom column"))
				}

			} else {
				cn = cn[1:]
				if typ.def == Def_Custom {
					for columnIdx++; columnIdx < nColumn; columnIdx++ {
						scn := nameRow[columnIdx]
						ln := len(scn)
						if ln == 0 {
							panic(fmt.Errorf("empty field name"))
						}

						col := &Column{csi: columnIdx}
						col.typ = toGoType(typeRow[columnIdx])
						if col.typ.def == Def_Custom {
							currRowIdx = rowIdx + RowHeader_Type
							panic(fmt.Errorf("field type can not be custom"))
						}

						typ.sub = append(typ.sub, col)
						e := ln - 1
						if scn[e] == Sign_ArrayEnd {
							col.name = scn[:e]
							break
						} else {
							col.name = scn
						}
					}

					var ei int
					err, ei = typ.InitCustomType()
					if err != nil {
						if ei == 0 {
							columnIdx = csi
						} else {
							columnIdx = ei
						}

						panic(err)
					}
				} else {
					typ.an = 1
					for columnIdx++; columnIdx < nColumn; columnIdx++ {
						scn := nameRow[columnIdx]
						ln := len(scn)
						if ln == 0 {
							panic(fmt.Errorf("empty name"))
						}

						e := ln - 1
						var end bool
						if scn[e] == Sign_ArrayEnd {
							scn = scn[:e]
							if scn == "" {
								panic(fmt.Errorf("empty name"))
							}
							end = true
						}

						if scn != cn {
							panic(fmt.Errorf("mismatch name, %s != %s", scn, cn))
						}

						if typeRow[columnIdx] != tn {
							currRowIdx = rowIdx + RowHeader_Type
							panic(fmt.Errorf("mismatch type, %s != %s", typeRow[columnIdx], tn))
						}

						typ.an++
						if end {
							break
						}
					}
				}
			}
		}

		currRowIdx = rowIdx + RowHeader_Type
		tb.cs = append(tb.cs, &Column{name: cn, typ: typ, csi: csi})
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
	if typ == Def_Custom {
		return &GoType{def: Def_Custom, sub: make([]*Column, 0, 4)}
	}

	gt, ok := simpleTypeMapping[typ]
	if !ok {
		if len(typ) > 2 && typ[0] == Sign_Ref {
			gt = toGoType(Def_PK)
			gt.ref = typ[1:]
			return gt
		}

		panic(fmt.Errorf("invalid type %s", typ))
	}

	return gt
}

func (inst *GoType) InitCustomType() (error, int) {
	m := make(map[string]int32, 4)
	ts := make([]reflect.Type, 0, 4)
	for _, col := range inst.sub {
		if col.name == "" {
			return fmt.Errorf("empty field name"), col.csi
		}

		v := m[col.name]
		if v == 0 {
			ts = append(ts, col.typ.typ)
		}

		m[col.name] = v + 1
	}

	var v0 int32
	for n, v1 := range m {
		if v0 == 0 {
			v0 = v1
			continue
		}

		if v1 != v0 {
			return fmt.Errorf("mismatch field %s", n), 0
		}
	}

	inst.an = int(v0)
	inst.typ = ctype.CustomType(ts)
	if inst.typ == nil {
		return fmt.Errorf("not found custom type"), 0
	}

	return nil, 0
}
