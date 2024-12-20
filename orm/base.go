package orm

import (
	"fmt"
	"reflect"

	"go-learner/orm/cache"
	"go-learner/orm/driver"

	"github.com/gogo/protobuf/proto"
)

type KeyFunc func(...any) string

type CacheFunc func(string, ...any) cache.IRecordCache

type T_DirtyFlag uint64

var typeMapping = make(map[string]*RecordType, 2048)

type RecordType struct {
	Type         reflect.Type
	Driver       driver.IDriver
	Id           int32
	NewCacheFunc func() cache.IRecordCache
	DecodeFunc   func([][]byte, proto.Message) (interface{}, error)
}

type BaseRecord struct {
	dirty    T_DirtyFlag
	raw_data [][]byte
}

func Get2[T, K any](key K) *T {
	var t *T
	tn := reflect.TypeOf(t).Elem().String()
	typ := typeMapping[tn]
	if typ == nil {
		panic(fmt.Errorf("not found record type for %s", tn))
	}

	return t
}

func (inst *RecordType) NewCache() cache.IRecordCache {

}

type Header struct {
	dirty T_DirtyFlag
}

func (inst *Header) SetDirty(index int) {
	inst.dirty = inst.dirty.Set(index)
}

func (v T_DirtyFlag) Set(index int) T_DirtyFlag {
	v &= 1 << index
	return v
}
