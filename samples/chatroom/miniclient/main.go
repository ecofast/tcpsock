package main

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ecofast/tcpsock"

	. "github.com/ecofast/tcpsock/samples/chatroom/protocol"
)

const (
	ServerAddr = "139.129.96.130:9999"

	charTableLen = 62
	charTable    = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

var (
	shutdown = make(chan bool, 1)

	tcpConn  *tcpsock.TcpConn
	userName string
)

func init() {
	rand.Seed(time.Now().UnixNano())

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signals
		shutdown <- true
	}()
}

func main() {
	genUserName()

	client := tcpsock.NewTcpClient(ServerAddr, onConnect, onClose, onProtocol)
	go client.Run()
	go input()
	<-shutdown
	client.Close()
}

func onProtocol() tcpsock.Protocol {
	proto := &ChatProtocol{}
	proto.OnMessage(onMsg)
	return proto
}

func onConnect(c *tcpsock.TcpConn) {
	log.Println("successfully connect to server", c.RawConn().RemoteAddr().String())
	tcpConn = c
}

func onClose(c *tcpsock.TcpConn) {
	log.Println("disconnect from server", c.RawConn().RemoteAddr().String())
	tcpConn = nil
}

func onMsg(c *tcpsock.TcpConn, p *ChatPacket) {
	fmt.Println(p)
}

func genUserName() {
	var buf bytes.Buffer
	for i := 0; i < 8; i++ {
		buf.WriteByte(charTable[rand.Intn(charTableLen)])
	}
	userName = buf.String()

	fmt.Println("your random name is:", userName)
}

func input() {
	s := ""
	for {
		if n, err := fmt.Scan(&s); n == 0 || err != nil {
			break
		}
		if tcpConn == nil {
			break
		}
		tcpConn.Write(genPacket(s))
	}
}

func genPacket(s string) *ChatPacket {
	var head PacketHead
	head.Signature = ChatSignature
	copy(head.UserName[:], []byte(userName))
	body := []byte(s)
	head.BodyLen = uint32(len(body))
	copy(body[:], body[:])
	return NewChatPacket(head, body)
}
