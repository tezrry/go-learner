package encoding

import "encoding/binary"

func Int64VariantEncode(v int64) []byte {
	rlt := make([]byte, 8)
	binary.PutVarint(rlt, v)
	return rlt
}
