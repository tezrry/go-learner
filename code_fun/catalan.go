package code_fun

import "fmt"

type BraceN struct {
	N   int
	num int
	buf []byte
}

func NewBraceN(n int) *BraceN {
	if n <= 1 {
		panic(fmt.Errorf("n MUST be greater than 1"))
	}

	return &BraceN{
		N:   n,
		buf: make([]byte, n<<1),
	}
}

func (inst *BraceN) PrintAll() {
	inst.fillBrace(inst.N, 0, 0)
}

func (inst *BraceN) fillBrace(leftMax, rightMax int, start int) {
	if start == len(inst.buf) {
		inst.printByteSlice()
		return
	}

	if leftMax > 0 {
		inst.buf[start] = '{'
		inst.fillBrace(leftMax-1, rightMax+1, start+1)
	}

	if rightMax > 0 {
		inst.buf[start] = '}'
		inst.fillBrace(leftMax, rightMax-1, start+1)
	}
}

func (inst *BraceN) printByteSlice() {
	e := inst.N << 1
	inst.num++
	fmt.Printf("%d: ", inst.num)
	for i := 0; i < e; i++ {
		fmt.Printf("%c", inst.buf[i])
	}
	fmt.Println()
}
