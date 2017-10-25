// Copyright (C) 2017 ecofast(胡光耀). All rights reserved.
// Use of this source code is governed by a BSD-style license.

package tcpsock

import (
	"net"
	"sync"
	"sync/atomic"
	"time"

	. "github.com/ecofast/rtl/sysutils"
)

type TcpServer struct {
	listener      *net.TCPListener
	acceptTimeout int
	*tcpSock
	autoIncID uint32
	numOfConn uint32
}

func NewTcpServer(listenPort, acceptTimeout int, onConnConnect, onConnClose OnTcpConnCallback, onCustomProtocol OnTcpCustomProtocol) *TcpServer {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ":"+IntToStr(int(listenPort)))
	CheckError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	CheckError(err)

	if onCustomProtocol == nil {
		panic("tcpsock.NewTcpServer: invalid custom protocol")
	}

	return &TcpServer{
		listener:      listener,
		acceptTimeout: acceptTimeout,
		tcpSock: &tcpSock{
			sendBufCap:       SendBufCapMax,
			recvBufCap:       RecvBufCapMax,
			exitChan:         make(chan struct{}),
			waitGroup:        &sync.WaitGroup{},
			onConnConnect:    onConnConnect,
			onConnClose:      onConnClose,
			onCustomProtocol: onCustomProtocol,
		},
	}
}

func (self *TcpServer) Serve() {
	self.waitGroup.Add(1)
	defer func() {
		self.listener.Close()
		self.waitGroup.Done()
	}()

	for {
		select {
		case <-self.exitChan:
			return

		default:
		}

		self.listener.SetDeadline(time.Now().Add(time.Duration(self.acceptTimeout) * time.Second))
		conn, err := self.listener.AcceptTCP()
		if err != nil {
			continue
		}

		atomic.AddUint32(&self.numOfConn, 1)
		self.waitGroup.Add(1)
		go func() {
			c := newTcpConn(atomic.AddUint32(&self.autoIncID, 1), self.tcpSock, conn, self.sendBufCap, self.recvBufCap, self.onCustomProtocol(), self.connClose)
			if self.onConnConnect != nil {
				self.onConnConnect(c)
			}
			c.run()
			self.waitGroup.Done()
		}()
	}
}

func (self *TcpServer) Close() {
	close(self.exitChan)
	self.waitGroup.Wait()
}

func (self *TcpServer) NumOfConn() uint32 {
	return atomic.LoadUint32(&self.numOfConn)
}

func (self *TcpServer) connClose(conn *TcpConn) {
	atomic.AddUint32(&self.numOfConn, ^uint32(0))
	if self.onConnClose != nil {
		self.onConnClose(conn)
	}
}
