package orm

import (
	"fmt"
	"reflect"
)

func Get[T any](ctx *Context, keys ...any) *T {
	var t *T
	tn := reflect.TypeOf(t).Elem().String()
	rt := typeMapping[tn]
	if rt == nil {
		panic(fmt.Errorf("not found record type for %s", tn))
	}

	cache := ctx.cache(rt)
	if cache == nil {
		cache = rt.NewCache()
		ctx.addCache(cache)
	}

	inst := cache.Get(keys)
	if inst != nil {
		return inst.(*T)
	}

	data := rt.Driver.Get(rt, keys)

	return t
}

func GetPartial[T any](ctx *Context, fields FieldFlag, keys ...any) *T {

}
