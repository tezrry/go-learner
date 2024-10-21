package hash

import (
	"hash/fnv"
	"testing"

	"go-learner/slice"

	"github.com/cespare/xxhash/v2"
	"github.com/spaolacci/murmur3"
	"github.com/zhenjl/cityhash"
)

var s = "fadfkas;dkgadfks;dfkopfk;"
var hv uint64

func Benchmark_fnv(b *testing.B) {
	bs := slice.String2ByteSlice(s)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h := fnv.New64a()
		h.Write(bs)
		hv = h.Sum64()
	}
}

func Benchmark_xxHash(b *testing.B) {
	bs := slice.String2ByteSlice(s)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hv = xxhash.Sum64(bs)
	}
}

func Benchmark_cityHash(b *testing.B) {
	bs := slice.String2ByteSlice(s)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hv = cityhash.CityHash64(bs, uint32(len(bs)))
	}
}

func Benchmark_murmurHash(b *testing.B) {
	bs := slice.String2ByteSlice(s)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h := murmur3.New64()
		h.Write(bs)
		hv = h.Sum64()
	}
}
