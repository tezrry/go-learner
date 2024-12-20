package cache

type IRecordCache interface {
	Get(keys ...any) any
	Set(t any, keys ...any)
	TypeId() int32
}

type List struct {
}
