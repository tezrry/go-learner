package reflect

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

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
