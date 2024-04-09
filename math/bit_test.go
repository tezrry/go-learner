package math

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFillBit(t *testing.T) {
	var i uint64 = 0b00010100
	fmt.Printf("%#b\n", fillBits(i))

	i = 0b100
	fmt.Printf("%#b\n", fillBits(i))

	i = 0
	fmt.Printf("%#b\n", fillBits(i))
}

func TestNumOf1(t *testing.T) {
	require.Equal(t, 0, NumOf1(0))
	require.Equal(t, 1, NumOf1(1))
	require.Equal(t, 2, NumOf1(0b00010100))
	require.Equal(t, 4, NumOf1(0b01010101))
	require.Equal(t, 6, NumOf1(0b11011101))
}

func TestClearBit(t *testing.T) {
	// &^ 二元运算符的操作结果是“bit clear"
	// a &^ b 的意思就是 清零a中，ab都为1的位
	t.Log(fmt.Sprintf("%#b", 0b0110&^0b1011)) // 0100
	t.Log(fmt.Sprintf("%#X", 0b1011&^0b1101)) // 0010
}

func TestNegativeBin(t *testing.T) {
	var x = int8(-6)
	// 补码 = 6: 0b00000110 -> 取反：0b11111001 -> +1: 0b11111010
	// 补码的补码 = 0b11111010 -> 0b00000101 -> 0b00000110 = 6
	// 相当于负负得正
	t.Log(fmt.Sprintf("%b", x))

	require.Equal(t, -x, ^x+1)
}
