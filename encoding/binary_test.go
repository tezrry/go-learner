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
