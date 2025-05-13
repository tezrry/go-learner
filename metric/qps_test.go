package metric

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestQPS(t *testing.T) {
	m := NewQPS("test")
	var wg sync.WaitGroup

	//m.Collect(time.Now().UnixMilli())

	nCPU := runtime.NumCPU()
	for i := 0; i < nCPU; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			timer := time.NewTimer(time.Second * 10)
			for {
				select {
				case <-timer.C:
					return
				default:
					time.Sleep(100 * time.Millisecond)
					m.Incr(1)
				}
			}
		}()
	}

	//n := 10000
	//m.Collect(time.Now().UnixMilli())
	//for i := 0; i < n; i++ {
	//	wg.Add(1)
	//	go func() {
	//		defer wg.Done()
	//		time.Sleep(10 * time.Millisecond)
	//		m.Incr(1)
	//	}()
	//}

	c := 10
	for i := 0; i < c; i++ {
		time.Sleep(1 * time.Second)
		v := m.Collect(time.Now().UnixMilli())
		t.Log(v)
	}

	wg.Wait()
}
