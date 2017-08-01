// Copyright (C) 2017 ecofast(胡光耀). All rights reserved.
// Use of this source code is governed by a BSD-style license.

// Package tcpsock provides easy to use interfaces for TCP I/O.
// Thanks to darksword(gansidui) and AlexStocks for their valuable projects
// which are gotcp(https://github.com/gansidui/gotcp)
// and getty(https://github.com/AlexStocks/getty).
package tcpsock

const (
	RecvBufLenMax = 4 * 1024
	SendBufLenMax = 4 * 1024

	cSendBufCap = 10
	cRecvBufCap = 10
)

type tcpSock struct {
	sendBufCap uint32
	recvBufCap uint32
	proto      Protocol
	exitChan   chan struct{}
}

type Protocol interface {
	Parse(b []byte, recvChan chan<- Packet)
	Process(conn *TcpConn, p Packet)
}

type Packet interface {
	Marshal() []byte
}
