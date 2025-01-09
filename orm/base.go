package orm

import (
	"reflect"

	"go-learner/orm/cache"

	"github.com/gogo/protobuf/proto"
)

type KeyFunc func(...any) string

type CacheFunc func(string, ...any) cache.IRecordCache

type FieldFlag uint64

var typeMapping = make(map[string]*RecordType, 2048)

type RecordType struct {
	DatabaseId   uint32
	TableId      uint32
	TypeName     string
	Type         reflect.Type
	Driver       IDriver
	NewCacheFunc func() cache.IRecordCache
	DecodeFunc   func([][]byte, proto.Message) (interface{}, error)
	KeyNum       uint16
}

type BaseRecord struct {
	dirty    FieldFlag
	active   FieldFlag
	raw_data [][]byte
}

func (inst *RecordType) NewCache() cache.IRecordCache {
	return inst.NewCacheFunc()
}

type Header struct {
	dirty FieldFlag
}

func (inst *Header) SetDirty(index int) {
	inst.dirty = inst.dirty.Set(index)
}

func (v FieldFlag) Set(index int) FieldFlag {
	v &= 1 << index
	return v
}
