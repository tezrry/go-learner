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

func newDataValue(x int) Data {
	return Data{x: x}
}
func newDataPointer(x int) *Data {
	return &Data{x: x}
}

func TestMethod(t *testing.T) {
	d := Data{}
	p := &d
	fmt.Printf("Data: %p ==>\n", p)

	d.ValueTest()   // ValueTest(d)
	d.PointerTest() // PointerTest(&d)

	p.ValueTest()   // ValueTest(*p)
	p.PointerTest() // PointerTest(p)

	newDataValue(1).ValueTest()
	// compiler fail, because it is not addressable
	//newDataValue(1).PointerTest()

	newDataPointer(1).ValueTest()
	newDataPointer(1).PointerTest()
}
