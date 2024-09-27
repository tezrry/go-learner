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
}

func TestClosure2(t *testing.T) {
	// 参数i在defer声明时已被求值
	for i := range whatever {
		defer func(ii int) { fmt.Println(ii) }(i)
	}
}

func TestClosure3(t *testing.T) {
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
		// 注意与作为参数的区别
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

func newTest() *test {
	return &test{"a"}
}
func panicFuncReturn(b bool, info string) (inst *test) {
	defer func() {
		if r := recover(); r != nil {
			inst = &test{info}
		}
	}()
	if b {
		panic("panicFunc")
	}
	return newTest()
}
func panicFunc(b bool, info string) (inst *test) {
	if b {
		panic("panicFunc")
	}
	return newTest()
}
func TestPanicReturn(t *testing.T) {
	var a, b *test
	defer func() {
		if a != nil {
			t.Log(a)
		}
		if b != nil {
			t.Log(b)
		}
	}()

	b = newTest()
	a = panicFuncReturn(true, "a panic")

	t.Log(a)
	t.Log(b)
}

func TestMultiPanic(t *testing.T) {
	var a *test
	defer func() {
		a = panicFuncReturn(true, "second panic")
		if r := recover(); r != nil {
			t.Log(r)
		}
		if r := recover(); r != nil {
			t.Log(r)
		}
	}()

	a = panicFunc(true, "first panic")
	t.Log(a)
}
