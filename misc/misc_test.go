package misc

import (
	"fmt"
	"testing"
)

func TestClearBit(t *testing.T) {
	// &^ 二元运算符的操作结果是“bit clear"
	// a &^ b 的意思就是 清零a中，ab都为1的位
	t.Log(fmt.Sprintf("%4b", 0b0110&^0b1011)) // 0100
	t.Log(fmt.Sprintf("%4b", 0b1011&^0b1101)) // 0010
}
