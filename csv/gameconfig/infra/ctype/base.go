package ctype

import (
	"unsafe"
)

type _ID uint32
type GID = _ID
type TID = _ID
type RID = _ID

const InvalidID = _ID(0)

type ENUM int32

type IDList []GID
type Params []float64

type Uint32List []uint32
type Float32List []float32

const (
	rowIdOffset = 18
	MaxRowId    = (1 << rowIdOffset) - 1
	MaxTableId  = (1 << (unsafe.Sizeof(TID(0))*8 - rowIdOffset)) - 1
	tableMask   = MaxTableId << rowIdOffset
)

func GlobalId(tableId TID, rowId RID) GID {
	return (tableId << rowIdOffset) | rowId
}

func TableId(gid GID) TID {
	return (gid & tableMask) >> rowIdOffset
}

func RowId(gid GID) RID {
	return gid & MaxRowId
}
