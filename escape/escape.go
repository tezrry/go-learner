package escape

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
