package slice

import (
	"unsafe"
)

func String2ByteSlice(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func ByteSlice2String(bs []byte) string {
	//return *((*string)(unsafe.Pointer(&bs)))
	return unsafe.String(unsafe.SliceData(bs), len(bs))
}
