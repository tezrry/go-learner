package orm

import (
	"context"

	"go-learner/orm/cache"
)

type Context struct {
	context.Context
	caches []cache.IRecordCache
}

func NewContext(ctx context.Context) *Context {
	return &Context{Context: ctx, caches: make([]cache.IRecordCache, 0, 8)}
}

func (inst *Context) cache(rt *RecordType) cache.IRecordCache {
	for _, rc := range inst.caches {
		if rc.TypeId() == rt.Id {
			return rc
		}
	}
	return nil
}

func (inst *Context) addCache(c cache.IRecordCache) {
	inst.caches = append(inst.caches, c)
}
