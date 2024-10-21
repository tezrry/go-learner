package sort

import (
	"slices"
	"testing"
)

type customT struct {
	i int
	v int
}

func sort(s []customT) {
	slices.SortFunc(s, func(a, b customT) int {
		return a.i - b.i
	})
}

func TestSlice(t *testing.T) {
	s := []customT{{i: 2, v: 2}, {i: 1, v: 1}}
	slices.SortFunc(s, func(a, b customT) int {
		return a.i - b.i
	})
	t.Log(s)

	s1 := []customT{{i: 2, v: 2}, {i: 1, v: 1}}
	sort(s1)
	t.Log(s1)
}
