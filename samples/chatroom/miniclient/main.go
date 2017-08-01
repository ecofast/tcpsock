package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"tcpsock"

	. "github.com/ecofast/sysutils"
)

const (
	ServerAddr = ":9999"

	ChatSignature  = 0xFFFFFFFF
	PacketHeadSize = 4 + 4 + 4
)

var (
	conn *net.TCPConn
	id   uint32

	recvChan = make(chan *ChatPacket)
	sendChan = make(chan string)
)

func main() {
	connToServer()
	genID()
	go run()
	go process()
	go input()

	select {}
}

func connToServer() {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ServerAddr)
	CheckError(err)
	conn, err = net.DialTCP("tcp", nil, tcpAddr)
	CheckError(err)
}

func genID() {
	fmt.Printf("pls enter your id: ")
	fmt.Scan(&id)
	fmt.Println("your id is:", id)
}

func process() {
	for {
		select {
		case p := <-recvChan:
			log.Printf("%d: %s\n", p.PlayerID, string(p.Body))
		case s := <-sendChan:
			conn.Write(genPacket(s).Marshal())
		}
	}
}

func input() {
	s := ""
	for {
		if n, err := fmt.Scan(&s); n == 0 || err != nil {
			break
		}
		sendChan <- s
	}
}

func run() {
	var head PacketHead
	var recvBuf []byte
	recvBufLen := 0
	buf := make([]byte, tcpsock.RecvBufLenMax)
	for {
		count, err := conn.Read(buf)
		if err != nil {
			break
		}

		if count+recvBufLen > tcpsock.RecvBufLenMax {
			continue
		}

		recvBuf = append(recvBuf, buf[0:count]...)
		recvBufLen += count
		offsize := 0
		offset := 0
		for recvBufLen-offsize > PacketHeadSize {
			offset = 0
			head.Signature = uint32(uint32(recvBuf[offsize+3])<<24 | uint32(recvBuf[offsize+2])<<16 | uint32(recvBuf[offsize+1])<<8 | uint32(recvBuf[offsize+0]))
			offset += 4
			head.PlayerID = uint32(uint32(recvBuf[offsize+offset+3])<<24 | uint32(recvBuf[offsize+offset+2])<<16 | uint32(recvBuf[offsize+offset+1])<<8 | uint32(recvBuf[offsize+offset+0]))
			offset += 4
			head.BodyLen = uint32(uint32(recvBuf[offsize+offset+3])<<24 | uint32(recvBuf[offsize+offset+2])<<16 | uint32(recvBuf[offsize+offset+1])<<8 | uint32(recvBuf[offsize+offset+0]))
			offset += 4
			if head.Signature == ChatSignature {
				pkglen := int(PacketHeadSize + head.BodyLen)
				if pkglen >= tcpsock.RecvBufLenMax {
					offsize = recvBufLen
					break
				}
				if offsize+pkglen > recvBufLen {
					break
				}

				pkt := NewChatPacket(head, recvBuf[offsize+offset:offsize+offset+int(head.BodyLen)])
				log.Printf("%d: %s\n", pkt.PlayerID, string(pkt.Body))

				offsize += pkglen
			} else {
				offsize++
			}
		}

		recvBufLen -= offsize
		if recvBufLen > 0 {
			recvBuf = recvBuf[offsize : offsize+recvBufLen]
		} else {
			recvBuf = nil
		}
	}
}

type PacketHead struct {
	Signature uint32
	PlayerID  uint32
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

func (p *ChatPacket) Marshal() []byte {
	buf := make([]byte, PacketHeadSize+len(p.Body))
	copy(buf[:PacketHeadSize], p.PacketHead.Bytes()[:])
	copy(buf[PacketHeadSize:], p.Body[:])
	return buf
}

func NewChatPacket(head PacketHead, body []byte) *ChatPacket {
	return &ChatPacket{
		PacketHead: head,
		Body:       body,
	}
}

func genPacket(s string) *ChatPacket {
	var head PacketHead
	head.Signature = ChatSignature
	head.PlayerID = id
	head.BodyLen = uint32(len(s))
	body := make([]byte, head.BodyLen)
	copy(body[:], []byte(s)[:])
	return NewChatPacket(head, body)
}
