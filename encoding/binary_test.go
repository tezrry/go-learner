package encoding

import (
	"encoding/binary"
	"testing"
)

func TestVariantInt64(t *testing.T) {
	buf := make([]byte, 8)
	n := binary.PutVarint(buf, -1)
	t.Log(n)
	t.Log(buf)
}

func TestNegative(t *testing.T) {
	i32 := int32(-10)
	ui32 := uint32(i32)

	i := int32(-3)
	i = int32(uint32(i) + ui32)
	t.Log(i)
}
