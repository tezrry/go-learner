package ctype

import (
	"unsafe"
)

type ID uint32
type TID = ID
type RID = ID

const InvalidID = ID(0)

type ENUM int32

type IDList []ID
type Params []float64

type Uint32List []uint32
type Float32List []float32

const (
	rowIdOffset = 18
	MaxRowId    = (1 << rowIdOffset) - 1
	MaxTableId  = (1 << (unsafe.Sizeof(TID(0))*8 - rowIdOffset)) - 1
	tableMask   = MaxTableId << rowIdOffset
)

func GlobalId(tableId TID, rowId RID) ID {
	return (tableId << rowIdOffset) | rowId
}

func TableId(gid ID) TID {
	return (gid & tableMask) >> rowIdOffset
}

func RowId(gid ID) RID {
	return gid & MaxRowId
}
