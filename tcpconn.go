package tcpsock

import (
	"net"
	"sync"
	"sync/atomic"
)

type TcpConn struct {
	id         uint32
	server     *TcpServer
	conn       *net.TCPConn
	sendChan   chan Packet
	recvChan   chan Packet
	closeChan  chan struct{}
	closeOnce  sync.Once
	closedFlag int32
}

func newTcpConn(id uint32, server *TcpServer, conn *net.TCPConn, sendCap, recvCap uint32) *TcpConn {
	return &TcpConn{
		id:        id,
		server:    server,
		conn:      conn,
		sendChan:  make(chan Packet, sendCap),
		recvChan:  make(chan Packet, recvCap),
		closeChan: make(chan struct{}),
	}
}

func (self *TcpConn) ID() uint32 {
	return self.id
}

func (self *TcpConn) run() {
	startGoroutine(self.reader, self.server.waitGroup)
	startGoroutine(self.writer, self.server.waitGroup)
	startGoroutine(self.handler, self.server.waitGroup)
}

func (self *TcpConn) Close() {
	self.closeOnce.Do(func() {
		atomic.StoreInt32(&self.closedFlag, 1)
		close(self.sendChan)
		close(self.recvChan)
		close(self.closeChan)
		self.conn.Close()
		self.server.connClose(self)
	})
}

func (self *TcpConn) Closed() bool {
	return atomic.LoadInt32(&self.closedFlag) == 1
}

func (self *TcpConn) RawConn() *net.TCPConn {
	return self.conn
}

func startGoroutine(fn func(), wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		fn()
		wg.Done()
	}()
}

func (self *TcpConn) reader() {
	defer func() {
		recover()
		self.Close()
	}()

	buf := make([]byte, RecvBufLenMax)
	for {
		select {
		case <-self.server.exitChan:
			return

		case <-self.closeChan:
			return

		default:
		}

		count, err := self.conn.Read(buf)
		if err != nil {
			return
		}
		self.server.proto.Parse(buf[:count], self.recvChan)
	}
}

func (self *TcpConn) writer() {
	defer func() {
		recover()
		self.Close()
	}()

	for {
		if self.Closed() {
			return
		}

		select {
		case <-self.server.exitChan:
			return

		case <-self.closeChan:
			return

		case p := <-self.sendChan:
			if _, err := self.conn.Write(p.Marshal()); err != nil {
				return
			}
		}
	}
}

func (self *TcpConn) handler() {
	defer func() {
		recover()
		self.Close()
	}()

	for {
		if self.Closed() {
			return
		}

		select {
		case <-self.server.exitChan:
			return

		case <-self.closeChan:
			return

		case packet := <-self.recvChan:
			self.server.proto.Process(self, packet)
		}
	}
}

func (self *TcpConn) Send(p Packet) {
	if self.Closed() {
		return
	}

	defer func() {
		recover()
	}()

	self.sendChan <- p
}
