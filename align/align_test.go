package align

import (
	"testing"
	"unsafe"
)

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
