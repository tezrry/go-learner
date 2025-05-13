package metric

import (
	"sync"
	"testing"
	"time"
)

func TestQPS(t *testing.T) {
	m := NewQPS("test")
	var wg sync.WaitGroup

	n := 10000
	m.Collect(time.Now().UnixMilli())
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(10 * time.Millisecond)
			m.Incr(1)
		}()
	}

	c := 3
	for i := 0; i < c; i++ {
		v := m.Collect(time.Now().UnixMilli())
		t.Log(v)
		time.Sleep(100 * time.Millisecond)
	}

	wg.Wait()
}
