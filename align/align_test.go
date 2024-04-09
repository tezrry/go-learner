package align

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
)

func TestAlign(t *testing.T) {
	require.Equal(t, uintptr(8), Align(1))
	require.Equal(t, uintptr(72), Align(66))
}

func TestName(t *testing.T) {
	type tAlign struct {
		i int16
		c byte
	}

	var ta tAlign

	t.Log(unsafe.Alignof(ta))
	t.Log(unsafe.Sizeof(ta))

	type tAlign2 struct {
		c byte
		i int16
	}

	var ta2 tAlign2
	t.Log(unsafe.Alignof(ta2))
	t.Log(unsafe.Sizeof(ta2))
}

func TestEmptyStruct(t *testing.T) {
	// struct{}作为结构体的最后一个成员时，需要占据空间。
	// 这是为了避免&demo3.a这种指针会指向结构体之外。
	type demo3 struct {
		c int32
		a struct{}
	}

	type demo4 struct {
		a struct{}
		c int32
	}

	var d3 demo3
	t.Log(unsafe.Alignof(d3))
	t.Log(unsafe.Sizeof(d3))

	var d4 demo4
	t.Log(unsafe.Alignof(d4))
	t.Log(unsafe.Sizeof(d4))
}
