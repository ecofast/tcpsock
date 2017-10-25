package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"tcpsock"
	"time"

	. "github.com/ecofast/rtl/timeutils"
)

const (
	ChatSignature  = 0xFFFFFFFF
	MaxUserNameLen = 8
	PacketHeadSize = 4 + MaxUserNameLen + 4
)

type PacketHead struct {
	Signature uint32
	UserName  [MaxUserNameLen]byte
	BodyLen   uint32
}

func (head *PacketHead) Bytes() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, head)
	return buf.Bytes()
}

type ChatPacket struct {
	PacketHead
	Body []byte
}

func NewChatPacket(head PacketHead, body []byte) *ChatPacket {
	return &ChatPacket{
		PacketHead: head,
		Body:       body,
	}
}

func (p *ChatPacket) Marshal() []byte {
	buf := make([]byte, PacketHeadSize+len(p.Body))
	copy(buf[:PacketHeadSize], p.PacketHead.Bytes()[:])
	copy(buf[PacketHeadSize:], p.Body[:])
	return buf
}

func (self *ChatPacket) String() string {
	return fmt.Sprintf("%s %s: %s", DateTimeToStr(time.Now()), string(self.UserName[:]), string(self.Body))
}

type ChatProtocol struct {
	recvBuf    []byte
	recvBufLen int
	onMsg      func(c *tcpsock.TcpConn, p *ChatPacket)
}

func (self *ChatProtocol) Parse(b []byte, recvChan chan<- tcpsock.Packet) {
	count := len(b)
	if count+self.recvBufLen > tcpsock.RecvBufLenMax {
		return
	}

	self.recvBuf = append(self.recvBuf, b[0:count]...)
	self.recvBufLen += count
	offsize := 0
	offset := 0
	var head PacketHead
	for self.recvBufLen-offsize > PacketHeadSize {
		offset = 0
		head.Signature = uint32(uint32(self.recvBuf[offsize+3])<<24 | uint32(self.recvBuf[offsize+2])<<16 | uint32(self.recvBuf[offsize+1])<<8 | uint32(self.recvBuf[offsize+0]))
		offset += 4
		copy(head.UserName[:], self.recvBuf[offsize+offset:offsize+offset+MaxUserNameLen])
		offset += MaxUserNameLen
		head.BodyLen = uint32(uint32(self.recvBuf[offsize+offset+3])<<24 | uint32(self.recvBuf[offsize+offset+2])<<16 | uint32(self.recvBuf[offsize+offset+1])<<8 | uint32(self.recvBuf[offsize+offset+0]))
		offset += 4
		if head.Signature == ChatSignature {
			pkglen := int(PacketHeadSize + head.BodyLen)
			if pkglen >= tcpsock.RecvBufLenMax {
				offsize = self.recvBufLen
				break
			}
			if offsize+pkglen > self.recvBufLen {
				break
			}

			recvChan <- NewChatPacket(head, self.recvBuf[offsize+offset : offsize+offset+int(head.BodyLen)][:])
			offsize += pkglen
		} else {
			offsize++
		}
	}

	self.recvBufLen -= offsize
	if self.recvBufLen > 0 {
		self.recvBuf = self.recvBuf[offsize : offsize+self.recvBufLen]
	} else {
		self.recvBuf = nil
	}
}

func (self *ChatProtocol) Process(conn *tcpsock.TcpConn, p tcpsock.Packet) {
	packet := p.(*ChatPacket)
	self.onMsg(conn, packet)
}

func (self *ChatProtocol) OnMessage(fn func(c *tcpsock.TcpConn, p *ChatPacket)) {
	self.onMsg = fn
}
