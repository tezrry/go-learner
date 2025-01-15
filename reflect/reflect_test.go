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

func TestGet(t *testing.T) {
	x := getInst[*testRecord]()
	t.Log(x.Int32())

	x2 := getInst[testRecord2]()
	t.Log(x2.Int32())
}

func TestTypeInfo(t *testing.T) {
	typ := reflect.TypeOf(uint8(0))
	t.Log(typ.String())
	t.Log(typ.Align())
	t.Log(typ.Size())
	t.Log(typ.FieldAlign())

	typ2 := reflect.TypeOf(any(nil))
	t.Log(typ2.String())
	t.Log(typ2.Align())
	t.Log(typ2.Size())
	t.Log(typ2.FieldAlign())

	typ3 := reflect.TypeOf([]int{})
	t.Log(typ3.String())
	t.Log(typ3.Align())
	t.Log(typ3.Size())
	t.Log(typ3.FieldAlign())

	typ4 := reflect.TypeOf(testRecord{})
	t.Log(typ4.String())
	t.Log(typ4.Align())
	t.Log(typ4.Size())
	t.Log(typ4.FieldAlign())
}
