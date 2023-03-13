package iface

import (
	"fmt"
	"testing"
)

type Mover interface {
	Move()
}

type dog struct{ int }

func (d *dog) Move() {
	fmt.Printf("dog: %p move\n", d)
}

type cat struct{ int }

func (c cat) Move() {
	fmt.Printf("cat: %p move\n", &c)
}

func Move(m Mover) {
	m.Move()
	m.Move()
}

func TestReceiver(t *testing.T) {
	pCat := &cat{1}
	fmt.Printf("cat %p ready ==>\n", pCat)
	Move(pCat)
	cat := cat{2}
	fmt.Printf("cat %p ready ==>\n", &cat)
	Move(cat)

	pDog := &dog{1}
	fmt.Printf("dog %p ready ==>\n", pDog)
	Move(pDog)
	dog := dog{2}
	fmt.Printf("dog %p ready ==>\n", &dog)
	// compile fail
	// Move(dog)
	Move(&dog)

	// 类型 T 方法集包含全部 receiver T 方法。
	// 类型 *T 方法集包含全部 receiver T + *T 方法。

	// 如类型 S 包含匿名字段 T，则 S 和 *S 方法集包含 T 方法。
	// 如类型 S 包含匿名字段 *T，则 S 和 *S 方法集包含 T + *T 方法。
	// 不管嵌入 T 或 *T，*S 方法集总是包含 T + *T 方法。
}
