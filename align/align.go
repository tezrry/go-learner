package align

import "math/bits"

const alignOf uintptr = bits.UintSize >> 3
const mask = ^(alignOf - 1)

// Align alignOf是2的n次幂，可以整除它的数，低n位必然为0。
// mask是对（alignOf-1）取反，&操作即为置低n位为0。
func Align(ptr uintptr) uintptr {
	return (ptr + alignOf - 1) & mask
}
