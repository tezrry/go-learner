package orm

type Header struct {
	dirty uint64
}

func (inst *Header) SetDirty(idx int) {
	inst.dirty &= 1 << idx
}
