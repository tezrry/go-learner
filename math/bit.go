package math

func IsPowerOfTwo(n uint64) bool {
	if n == 0 {
		return false
	}

	return clearLast1(n) == 0
}

// CeilToPowerOfTwo returns the least power of two integer value greater than
// or equal to n.
func CeilToPowerOfTwo(n uint64) uint64 {
	if n <= 2 {
		return n
	}
	n--
	n = fillBits(n)
	n++
	return n
}

// FloorToPowerOfTwo returns the greatest power of two integer value less than
// or equal to n.
func FloorToPowerOfTwo(n uint64) uint64 {
	if n <= 2 {
		return n
	}
	n = fillBits(n)
	n >>= 1
	n++
	return n
}

// NumOf1 二进制表示中1的个数
func NumOf1(n uint64) int {
	if n == 0 {
		return 0
	}

	var num = 0
	for {
		n = clearLast1(n)
		num++

		if n == 0 {
			break
		}
	}

	return num
}

// 去掉二进制表示的最后一个1，比如: 0b010100 -> 0b010000
func clearLast1(n uint64) uint64 {
	return n & (n - 1)
}

// 将二进制表示的第一个1后的所有0都填充为1，例如: 0b0100 -> 0b0111
// 思路为：
// 先按照每1位分组，组1 ｜ 组2，组3 ｜ 组4，...
// 再按照每2位分组，组1 ｜ 组2，组3 ｜ 组4，...
// 再按照每4位分组，组1 ｜ 组2，组3 ｜ 组4，...
// 循环到按照每32位分组，此时只会分为2组。
func fillBits(n uint64) uint64 {
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32
	return n
}
