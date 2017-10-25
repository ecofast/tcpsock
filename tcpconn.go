// Copyright (C) 2017 ecofast(胡光耀). All rights reserved.
// Use of this source code is governed by a BSD-style license.

package tcpsock

import (
	"net"
	"sync"
	"sync/atomic"
)

type OnTcpConnCallback func(c *TcpConn)

type OnTcpCustomProtocol func() Protocol

type TcpConn struct {
	id         uint32
	owner      *tcpSock
	conn       *net.TCPConn
	sendChan   chan Packet
	recvChan   chan Packet
	proto      Protocol
	closeChan  chan struct{}
	closeOnce  sync.Once
	closedFlag int32
	onClose    OnTcpConnCallback
}

func newTcpConn(id uint32, owner *tcpSock, conn *net.TCPConn, sendCap, recvCap uint32, proto Protocol, onClose OnTcpConnCallback) *TcpConn {
	return &TcpConn{
		id:        id,
		owner:     owner,
		conn:      conn,
		sendChan:  make(chan Packet, sendCap),
		recvChan:  make(chan Packet, recvCap),
		proto:     proto,
		closeChan: make(chan struct{}),
		onClose:   onClose,
	}
}

func (self *TcpConn) ID() uint32 {
	return self.id
}

func (self *TcpConn) run() {
	startGoroutine(self.reader, self.owner.waitGroup)
	startGoroutine(self.writer, self.owner.waitGroup)
	startGoroutine(self.handler, self.owner.waitGroup)
}

func (self *TcpConn) Close() {
	self.closeOnce.Do(func() {
		atomic.StoreInt32(&self.closedFlag, 1)
		close(self.sendChan)
		close(self.recvChan)
		close(self.closeChan)
		self.conn.Close()
		if self.onClose != nil {
			self.onClose(self)
		}
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
		case <-self.owner.exitChan:
			return

		case <-self.closeChan:
			return

		default:
		}

		count, err := self.conn.Read(buf)
		if err != nil {
			return
		}
		self.proto.Parse(buf[:count], self.recvChan)
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
		case <-self.owner.exitChan:
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
		case <-self.owner.exitChan:
			return

		case <-self.closeChan:
			return

		case packet := <-self.recvChan:
			self.proto.Process(self, packet)
		}
	}
}

func (self *TcpConn) Write(p Packet) {
	if self.Closed() {
		return
	}

	defer func() {
		recover()
	}()

	self.sendChan <- p
}
