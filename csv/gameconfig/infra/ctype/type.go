package ctype

import "unsafe"

type _ID uint32
type GID = _ID
type TID = _ID
type RID = _ID

const InvalidID = _ID(0)

type ENUM int32

type IDInt32Pair struct {
	Id    GID
	Value int32
}

type IDUint32Pair struct {
	Id    GID
	Value uint32
}

type IDFloat32Pair struct {
	Id    GID
	Value float32
}

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

func (inst Params) IterateIDFloatPair(f func(IDFloat32Pair)) {
	lp := len(inst) >> 1
	for i := 0; i < lp; i++ {
		ii := i << 1
		f(IDFloat32Pair{GID(inst[ii]), float32(inst[ii+1])})
	}
}
