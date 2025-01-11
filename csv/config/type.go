package config

import "unsafe"

type ID uint32

type IDInt32Pair struct {
	Id    ID
	Value int32
}

type IDUint32Pair struct {
	Id    ID
	Value uint32
}

type IDFloat32Pair struct {
	Id    ID
	Value float32
}

type IDList []ID
type Params []float64

type Uint32List []uint32
type Float32List []float32

const (
	rowIdOffset = 18
	MaxRowId    = (1 << rowIdOffset) - 1
	MaxTableId  = (1 << (unsafe.Sizeof(ID(0))*8 - rowIdOffset)) - 1
	tableMask   = MaxTableId << rowIdOffset
)

func (inst Params) IterateIDFloatPair(f func(IDFloat32Pair)) {
	lp := len(inst) >> 1
	for i := 0; i < lp; i++ {
		ii := i << 1
		f(IDFloat32Pair{ID(inst[ii]), float32(inst[ii+1])})
	}
}
