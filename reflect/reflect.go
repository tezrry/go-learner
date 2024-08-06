package reflect

import (
	"fmt"
	"reflect"
)

type meta struct {
	t reflect.Type
}

type testRecord struct {
	i   int32
	s   string
	f   float64
	ptr *testRecord
}

func printReflection(itf interface{}) {
	typ := reflect.TypeOf(itf)
	fmt.Printf("type = %v\n", typ)
}
