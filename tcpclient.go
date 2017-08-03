// Copyright (C) 2017 ecofast(胡光耀). All rights reserved.
// Use of this source code is governed by a BSD-style license.

package tcpsock

import (
	"net"
	"sync"

	. "github.com/ecofast/sysutils"
)

type TcpClient struct {
	svrAddr *net.TCPAddr
	*tcpSock
}

func NewTcpClient(svrAddr string, proto Protocol) *TcpClient {
	tcpAddr, err := net.ResolveTCPAddr("tcp", svrAddr)
	CheckError(err)
	return &TcpClient{
		svrAddr: tcpAddr,
		tcpSock: &tcpSock{
			sendBufCap: SendBufCapMax,
			recvBufCap: RecvBufCapMax,
			proto:      proto,
			exitChan:   make(chan struct{}),
			waitGroup:  &sync.WaitGroup{},
		},
	}
}

func (self *TcpClient) Run() {
	conn, err := net.DialTCP("tcp", nil, self.svrAddr)
	CheckError(err)

	self.waitGroup.Add(1)
	go func() {
		// client sock do NOT need to identify self
		c := newTcpConn( /*atomic.AddUint32(&self.autoIncID, 1)*/ 0, self.tcpSock, conn, self.sendBufCap, self.recvBufCap, self.connClose)
		if self.onConnConnect != nil {
			self.onConnConnect(c)
		}
		c.run()
		self.waitGroup.Done()
	}()
}

func (self *TcpClient) Close() {
	close(self.exitChan)
	self.waitGroup.Wait()
}

func (self *TcpClient) OnConnect(fn OnTcpConnCallback) {
	self.onConnConnect = fn
}

func (self *TcpClient) OnClose(fn OnTcpConnCallback) {
	self.onConnClose = fn
}

func (self *TcpClient) connClose(conn *TcpConn) {
	if self.onConnClose != nil {
		self.onConnClose(conn)
	}
}
