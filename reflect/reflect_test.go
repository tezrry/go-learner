package reflect

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

type testStruct struct {
	i int
}

func (t *testStruct) Method1(i int) int {
	return t.i + i
}

func (t *testStruct) method1(i *int) bool {
	return true
}

func TestName(t *testing.T) {
	v := testRecord{i: 1}
	tv := reflect.TypeOf(v)
	vv := reflect.ValueOf(v)
	fmt.Printf("v.Type:%v\n", tv)
	fmt.Printf("v.Value:%v\n", vv)
	require.Panics(t, func() { tv.Elem() })
	require.Panics(t, func() { vv.Elem() })
	fmt.Printf("v.Value.Indirect:%v\n", reflect.Indirect(vv))
	require.Panics(t, func() { reflect.Indirect(vv).Addr() })

	p := &testRecord{i: 2}
	tp := reflect.TypeOf(p)
	vp := reflect.ValueOf(p)
	fmt.Printf("p.Type:%v\n", tp)
	fmt.Printf("p.Value:%v\n", vp)
	fmt.Printf("p.Value.elem:%v\n", vp.Elem())
	fmt.Printf("p.Value.elem.Type:%v\n", vp.Type().Elem())
	fmt.Printf("p.Value.Indirect:%v\n", reflect.Indirect(vp))
	fmt.Printf("p.Value.elem.Addr:%v\n", vp.Elem().Addr())
}

func TestString(t *testing.T) {
	v := &testRecord{i: 1}
	tv := reflect.TypeOf(v)
	t.Log(tv.Name())
	t.Log(tv.String())
	t.Log(reflect.TypeOf((*testRecord)(nil)).String())
}

func TestIterateMethods(t *testing.T) {
	st := &testStruct{i: 1}
	tv := reflect.ValueOf(st)
	tt := tv.Type()
	for i := 0; i < tt.NumMethod(); i++ {
		t.Log(tt.Method(i).Name)
		mt := tv.Method(i).Type()
		for j := 0; j < mt.NumIn(); j++ {
			t.Logf("%v, %s", mt, mt.In(j).Name())
		}
		tv.Method(i).Call([]reflect.Value{reflect.ValueOf(2)})
	}
}
