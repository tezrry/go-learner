package ctype

import (
	"reflect"
	"slices"
	"strings"
)

var customMapping = make(map[string]reflect.Type, 16)

func init() {
	RegisterCustomType(reflect.TypeOf(IDInt32{}))
	RegisterCustomType(reflect.TypeOf(IDUint32{}))
	RegisterCustomType(reflect.TypeOf(IDFloat32{}))
	RegisterCustomType(reflect.TypeOf(Int32Float32String{}))
}

func RegisterCustomType(t reflect.Type) {
	if t.Kind() != reflect.Struct {
		panic("not struct")
	}

	num := t.NumField()
	sts := make([]reflect.Type, num)
	for i := 0; i < num; i++ {
		sts[i] = t.Field(i).Type
	}

	key := typeName(sts)
	_, ok := customMapping[key]
	if ok {
		panic("duplicate key " + key)
	}

	customMapping[key] = t
}

func CustomType(ts []reflect.Type) reflect.Type {
	return customMapping[typeName(ts)]
}

func typeName(ts []reflect.Type) string {
	slices.SortFunc(ts, func(a, b reflect.Type) int {
		rtn := int(a.Size() - b.Size())
		if rtn == 0 {
			if a.String() < b.String() {
				return -1
			}
			return 1
		}
		return rtn
	})

	var builder strings.Builder
	for _, t := range ts {
		builder.WriteString(t.String())
		builder.WriteByte('_')
	}

	return builder.String()
}

type IDInt32 struct {
	Id    ID
	Value int32
}

type IDUint32 struct {
	Id    ID
	Value uint32
}

type IDFloat32 struct {
	Id    ID
	Value float32
}

type Int32Float32String struct {
	I int32
	F float32
	S string
}
