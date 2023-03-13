package net

import (
	"fmt"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type handler func(conn net.Conn)
type tcpEchoServer struct {
	stopped atomic.Bool
	address string
	handler handler
}

func newTcpEchoServer(address string, handler handler) *tcpEchoServer {
	return &tcpEchoServer{
		address: address,
		handler: handler,
	}
}

func (srv *tcpEchoServer) start() {
	lis, err := net.Listen("tcp", srv.address)
	if err != nil {
		panic(err)
	}

	for !srv.stopped.Load() {
		conn, err := lis.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		go srv.handler(conn)
	}
}

func (srv *tcpEchoServer) stop() {
	srv.stopped.Store(true)
}

func TestTcpEchoServer(t *testing.T) {
	addr := ":6789"
	srv := newTcpEchoServer(addr, func(conn net.Conn) {
		nBuf := 4
		buf := make([]byte, nBuf)
		n := nBuf
		for n == nBuf {
			var err error
			n, err = conn.Read(buf)
			assert.NoError(t, err)

			_, err = conn.Write(buf[:n])
			assert.NoError(t, err)
		}

		err := conn.Close()
		assert.NoError(t, err)
	})
	go srv.start()
	time.Sleep(time.Second)

	conn, err := net.Dial("tcp", addr)
	assert.NoError(t, err)

	msg := "hello"
	t.Logf("send: %s", msg)
	nMsg := len(msg)
	_, err = conn.Write([]byte(msg))
	assert.NoError(t, err)

	n := 0
	buf := make([]byte, nMsg)
	n, err = conn.Read(buf)
	assert.Equal(t, n, nMsg)

	t.Logf("recv: %s", string(buf))
	srv.stop()
}
