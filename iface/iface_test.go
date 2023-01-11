package iface

import "testing"

type Mover interface {
	Move()
}

type dog struct{}

func (d *dog) Move() {
	println("dog move")
}

type cat struct{}

func (c cat) Move() {
	println("cat move")
}

func Move(m Mover) {
	m.Move()
}

func TestReceiver(t *testing.T) {
	pCat := &cat{}
	Move(pCat)
	Move(cat{})

	pDog := &dog{}
	Move(pDog)
	// compile fail
	//Move(dog{})
}
