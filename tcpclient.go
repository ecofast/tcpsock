// Copyright (C) 2017 ecofast(胡光耀). All rights reserved.
// Use of this source code is governed by a BSD-style license.

package tcpsock

import (
	"net"
	"sync"

	. "github.com/ecofast/rtl/sysutils"
)

type TcpClient struct {
	svrAddr *net.TCPAddr
	*tcpSock
}

func NewTcpClient(svrAddr string, onConnConnect, onConnClose OnTcpConnCallback, onCustomProtocol OnTcpCustomProtocol) *TcpClient {
	tcpAddr, err := net.ResolveTCPAddr("tcp", svrAddr)
	CheckError(err)

	if onCustomProtocol == nil {
		panic("tcpsock.NewTcpClient: invalid custom protocol")
	}

	return &TcpClient{
		svrAddr: tcpAddr,
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

func (self *TcpClient) Run() {
	conn, err := net.DialTCP("tcp", nil, self.svrAddr)
	CheckError(err)

	self.waitGroup.Add(1)
	go func() {
		c := newTcpConn(0, self.tcpSock, conn, self.sendBufCap, self.recvBufCap, self.onCustomProtocol(), self.connClose)
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

func (self *TcpClient) connClose(conn *TcpConn) {
	if self.onConnClose != nil {
		self.onConnClose(conn)
	}
}
