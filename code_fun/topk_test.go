package code_fun

import "testing"

func TestTopK(t *testing.T) {
	ret := TopK([]int{1, 2, 3, 4, 5, 10, 1, 2}, 2)
	t.Log(ret)
}
