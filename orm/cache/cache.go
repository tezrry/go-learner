package cache

type IRecordCache interface {
	Get(keys ...any) any
	Set(t any, keys ...any)
	TypeId() uint32
}

type List struct {
}
