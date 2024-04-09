package code_fun

import (
	"fmt"
	"testing"
)

func TestBraceN(t *testing.T) {
	N := 2
	t.Run(fmt.Sprintf("N_%d", N), func(t *testing.T) {
		bn := NewBraceN(N)
		bn.PrintAll()
	})

	N = 3
	t.Run(fmt.Sprintf("N_%d", N), func(t *testing.T) {
		bn := NewBraceN(N)
		bn.PrintAll()
	})

	N = 4
	t.Run(fmt.Sprintf("N_%d", N), func(t *testing.T) {
		bn := NewBraceN(N)
		bn.PrintAll()
	})

	N = 5
	t.Run(fmt.Sprintf("N_%d", N), func(t *testing.T) {
		bn := NewBraceN(N)
		bn.PrintAll()
	})
}
