package reflect

import (
	"fmt"
	"reflect"
)

type ITest interface {
	Int32() int32
}

type testRecord struct {
	i   int32
	s   string
	f   float64
	ptr *testRecord
}

type testRecord2 struct {
	i int32
}

func (inst *testRecord) Int32() int32 {
	return inst.i
}

func (inst testRecord2) Int32() int32 {
	return inst.i
}

func printReflection(itf interface{}) {
	typ := reflect.TypeOf(itf)
	fmt.Printf("type = %v\n", typ)
}

func getInst[T ITest]() T {
	var t T
	return t
}

func getName[T any]() string {
	var t *T
	//typ := reflect.TypeOf(t).
	return reflect.TypeOf(t).Elem().String()
}
