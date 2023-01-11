package _func

import (
	"fmt"
	"testing"
)

type Data struct {
	x int
}

func (d Data) ValueTest() { // func ValueTest(self Data);
	fmt.Printf("Value: %p\n", &d)
}

func (d *Data) PointerTest() { // func PointerTest(self *Data);
	fmt.Printf("Pointer: %p\n", d)
}

func TestMethod(t *testing.T) {
	d := Data{}
	p := &d
	fmt.Printf("Data: %p\n", p)

	d.ValueTest()   // ValueTest(d)
	d.PointerTest() // PointerTest(&d)

	p.ValueTest()   // ValueTest(*p)
	p.PointerTest() // PointerTest(p)
}
