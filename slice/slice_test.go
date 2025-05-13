package slice

import (
	"fmt"
	"strconv"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
)

func TestSlice_cap(t *testing.T) {
	s1 := []byte{0}
	// 初始化时没有用size class
	require.Equal(t, 1, len(s1))
	require.Equal(t, 1, cap(s1))

	b := make([]byte, 1)
	s1 = append(s1, b...)
	require.Equal(t, 2, len(s1))
	// min size class is 8 byte
	require.Equal(t, 8, cap(s1))

	b = make([]byte, 10)
	s1 = append(s1, b...)
	require.Equal(t, 12, len(s1))
	require.Equal(t, 16, cap(s1))

	b = make([]byte, 21)
	s1 = append(s1, b...)
	require.Equal(t, 33, len(s1))
	// 33 > 16*2，故用33去寻找合适的size class：48
	require.Equal(t, 48, cap(s1))
}

func TestSlice_append(t *testing.T) {
	s1 := []int{0}
	fmt.Printf("s1:%v, len=%d, cap=%d, ptr=%p\n", s1, len(s1), cap(s1), &s1[0])
	fmt.Println("=================")

	// s1的cap为1，故s2使用新的底层数组，cap为2
	s2 := append(s1, 1)
	fmt.Printf("s1:%v, len=%d, cap=%d, ptr=%p\n", s1, len(s1), cap(s1), &s1[0])
	fmt.Printf("s2:%v, len=%d, cap=%d, ptr:%p\n", s2, len(s2), cap(s2), &s2[0])
	fmt.Println("=================")

	// s2扩容，cap为4
	s2 = append(s2, 2)
	fmt.Printf("s2:%v, len=%d, cap=%d, ptr:%p\n", s2, len(s2), cap(s2), &s2[0])
	fmt.Println("=================")

	// s3使用和s2同样的底层数组
	s3 := append(s2, 3)
	fmt.Printf("s2:%v, len=%d, cap=%d, ptr:%p\n", s2, len(s2), cap(s2), &s2[0])
	fmt.Printf("s3:%v, len=%d, cap=%d, ptr:%p\n", s3, len(s3), cap(s3), &s3[0])
	fmt.Println("=================")

	// s4使用和s2同样的底层数组，此时会导致s3[3]从3变成4
	s4 := append(s2, 4)
	fmt.Printf("s2:%v, len=%d, cap=%d, ptr:%p\n", s2, len(s2), cap(s2), &s2[0])
	fmt.Printf("s3:%v, len=%d, cap=%d, ptr:%p\n", s3, len(s3), cap(s3), &s3[0])
	fmt.Printf("s4:%v, len=%d, cap=%d, ptr:%p\n", s4, len(s4), cap(s4), &s4[0])
}

func TestSlice_append2(t *testing.T) {
	s1 := make([]int, 0, 4)
	for i := 0; i < 3; i++ {
		s1 = append(s1, i)
	}

	s2 := append(s1, 3, 4)
	fmt.Printf("s1:%v, len=%d, cap=%d, ptr:%p\n", s1, len(s1), cap(s1), &s1[0])
	fmt.Printf("s2:%v, len=%d, cap=%d, ptr:%p\n", s2, len(s2), cap(s2), &s2[0])

	ps1_3 := (uintptr)((unsafe.Pointer)(&s1[2])) + 8
	fmt.Printf("s1[3] = %d\n", *((*int)((unsafe.Pointer)(ps1_3))))

	a1 := (*[4]int)((unsafe.Pointer)(&s1[0]))
	fmt.Printf("a1:%v\n", *a1)

	s3 := append(s1, 5)
	fmt.Printf("s1[3] = %d\n", *((*int)((unsafe.Pointer)(ps1_3))))
	fmt.Printf("a1:%v\n", *a1)
	fmt.Printf("s2:%v, len=%d, cap=%d, ptr:%p\n", s2, len(s2), cap(s2), &s2[0])
	fmt.Printf("s3:%v, len=%d, cap=%d, ptr:%p\n", s3, len(s3), cap(s3), &s3[0])
}

func TestByteSlice2String(t *testing.T) {
	bs := []byte{'a', 'b'}
	s := ByteSlice2String(bs)
	require.Equal(t, "ab", s)

	require.Equal(t,
		uintptr(unsafe.Pointer(unsafe.SliceData(bs))),
		uintptr(unsafe.Pointer(unsafe.StringData(s))))
}

func TestString2ByteSlice(t *testing.T) {
	s := "ab"
	bs := String2ByteSlice(s)
	require.Equal(t, 2, len(bs))
	require.Equal(t, 2, cap(bs))
	require.Equal(t, byte('a'), bs[0])
	require.Equal(t, byte('b'), bs[1])

	require.Equal(t,
		uintptr(unsafe.Pointer(unsafe.StringData(s))),
		uintptr(unsafe.Pointer(unsafe.SliceData(bs))))
}

func Benchmark_ExpandSlice_1(b *testing.B) {
	b1 := make([]byte, 1024)
	for i := 0; i < b.N; i++ {
		var b0 = []byte{'0'}
		b0 = append(b0, b1...)
	}
}

func Benchmark_ExpandSlice_2(b *testing.B) {
	b1 := make([]byte, 1024)
	for i := 0; i < b.N; i++ {
		var b0 = []byte{'0'}
		copy(b1, b0)
		b0 = b1
	}
}

func Benchmark_ExpandSlice_3(b *testing.B) {
	s1 := ByteSlice2String(make([]byte, 1024))
	for i := 0; i < b.N; i++ {
		var b0 = []byte{'0'}
		b0 = append(b0, s1...)
	}
}

func Benchmark_appendInt_1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var b0 = make([]byte, 0, 1024)
		b0 = strconv.AppendInt(b0, int64(i), 10)
	}
}

func Benchmark_appendInt_2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var b0 = make([]byte, 1024)
		s := strconv.FormatInt(int64(i), 10)
		copy(b0, s)
	}
}

func Benchmark_iterate_vs_map(b *testing.B) {
	n := 16
	s := make([]int, n)
	for i := 0; i < n; i++ {
		s[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, v := range s {
			if v == n+1 {
				b.Log("ok")
			}
		}
	}
}

func Benchmark_iterate_vs_map1(b *testing.B) {
	n := 16
	m := make(map[int]int, n)
	for i := 0; i < n; i++ {
		m[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, ok := m[n+1]; ok {
			b.Log("ok")
		}
	}
}

func TestSliceDelete(t *testing.T) {
	b := []int{0}
	b = append(b[:0], b[1:]...)
	t.Log(b)

	b = []int{0, 1, 2}
	b = append(b[:0], b[1:]...)
	t.Log(b)

	b = []int{0, 1, 2}
	b = append(b[:1], b[2:]...)
	t.Log(b)

	b = []int{0, 1, 2}
	b = append(b[:2], b[3:]...)
	t.Log(b)
}
