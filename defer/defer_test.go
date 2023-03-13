package _defer

import (
	"fmt"
	"testing"
)

var whatever [3]struct{}

func TestDeferFILO(t *testing.T) {
	for i := range whatever {
		defer fmt.Println(i)
	}
}

func TestClosure(t *testing.T) {
	// 变量i被闭包引用，defer执行时，输出的i是最后被修改的值
	for i := range whatever {
		defer func() { fmt.Println(i) }()
	}

	// 参数i在defer声明时已被求值
	for i := range whatever {
		defer func(ii int) { fmt.Println(ii) }(i)
	}
}

func TestClosure2(t *testing.T) {
	// 变量i作为闭包引用，但函数依次执行，defer也是依次执行
	for i := range whatever {
		func() {
			defer fmt.Println(i)
		}()
	}
}

type test struct {
	name string
}

func (t *test) Close() {
	fmt.Println(t.name, " closed")
}
func valueClose(t test) {
	t.Close()
}
func pointerClose(t *test) {
	t.Close()
}
func TestStructPointer(t *testing.T) {
	ts := []test{{"a"}, {"b"}, {"c"}}
	for _, t := range ts {
		defer t.Close()
	}
}
func TestStructPointer2(t *testing.T) {
	ts := []test{{"a"}, {"b"}, {"c"}}
	for _, t := range ts {
		t2 := t
		defer t2.Close()
	}
}
func TestStructPointer3(t *testing.T) {
	ts := []test{{"a"}, {"b"}, {"c"}}
	// t作为参数，被实时求值
	for _, t := range ts {
		defer valueClose(t)
	}
}
func TestStructPointer4(t *testing.T) {
	ts := []test{{"a"}, {"b"}, {"c"}}
	for _, t := range ts {
		// t作为参数，被实时求值，但&t指向的其实是同一个地址
		defer pointerClose(&t)
	}
}
func TestStructPointer5(t *testing.T) {
	ts := []*test{{"a"}, {"b"}, {"c"}}
	for _, t := range ts {
		// t作为参数，被实时求值
		defer pointerClose(t)
	}
}
