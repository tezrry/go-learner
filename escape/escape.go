package escape

// go build -gcflags="..."
// -m/-m -m 逃逸分析
// -l 禁止内联
// -N 禁止编译器优化
// -S 查看汇编
type foo struct {
	v int64
}

func paramWithPointer_01(p *foo) {
	//  p does not escape
	p.v = 1
}

func returnPointer_02(v int64) *foo {
	// &foo{...} escapes to heap
	return &foo{v: v}
}

func test_01() {
	var f foo
	paramWithPointer_01(&f)
}

func test_02() {
	// go build gcflags='-m'
	// inlining call to returnPointer_02
	// &foo{...} does not escape
	f := returnPointer_02(1)
	f.v = 0
}
