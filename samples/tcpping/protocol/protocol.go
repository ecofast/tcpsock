package protocol

import (
	"encoding/binary"

	"github.com/ecofast/tcpsock"
)

const (
	PacketHeadSize = 2
)

type PingPacket struct {
	BodyLen uint16
	Body    []byte
}

func NewPingPacket(body []byte) *PingPacket {
	return &PingPacket{
		BodyLen: uint16(len(body)),
		Body:    body,
	}
}

func (self *PingPacket) Marshal() []byte {
	buf := make([]byte, PacketHeadSize+self.BodyLen)
	binary.LittleEndian.PutUint16(buf, self.BodyLen)
	copy(buf[PacketHeadSize:], self.Body[:])
	return buf
}

type PingProtocol struct {
	recvBuf    []byte
	recvBufLen int
	onMsg      func(c *tcpsock.TcpConn, p *PingPacket)
}

func (self *PingProtocol) Parse(b []byte, recvChan chan<- tcpsock.Packet) {
	count := len(b)
	if count+self.recvBufLen > tcpsock.RecvBufLenMax {
		return
	}

	self.recvBuf = append(self.recvBuf, b[0:count]...)
	self.recvBufLen += count
	offsize := 0
	offset := 0
	var pkt PingPacket
	for self.recvBufLen-offsize > 2 {
		offset = 0
		pkt.BodyLen = binary.LittleEndian.Uint16(self.recvBuf[offsize+0 : offsize+2])
		offset += 2
		pkglen := int(2 + pkt.BodyLen)
		if pkglen >= tcpsock.RecvBufLenMax {
			offsize = self.recvBufLen
			break
		}
		if offsize+pkglen > self.recvBufLen {
			break
		}

		recvChan <- NewPingPacket(self.recvBuf[offsize+offset : offsize+offset+int(pkt.BodyLen)])
		offsize += pkglen
	}

	self.recvBufLen -= offsize
	if self.recvBufLen > 0 {
		self.recvBuf = self.recvBuf[offsize : offsize+self.recvBufLen]
	} else {
		self.recvBuf = nil
	}
}

func (self *PingProtocol) Process(conn *tcpsock.TcpConn, p tcpsock.Packet) {
	packet := p.(*PingPacket)
	self.onMsg(conn, packet)
}

func (self *PingProtocol) OnMessage(fn func(c *tcpsock.TcpConn, p *PingPacket)) {
	self.onMsg = fn
}
