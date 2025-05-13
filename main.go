package main

import "time"

func testPanic(msg string) {
	panic(msg)
}

func main() {
	t := time.NewTimer(time.Second * 30)
	defer t.Stop()

	currTime := time.Now()
	println(currTime.Format("2006-01-02 15:04:05"))
	select {
	case <-t.C:
		println(time.Now().Sub(currTime))
		println(time.Now().Format("2006-01-02 15:04:05"))
	}
}
