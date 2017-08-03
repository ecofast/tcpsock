package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"tcpsock"
	. "tcpsock/samples/chatroom/protocol"
)

const (
	ServerAddr = ":9999"
)

var (
	shutdown = make(chan bool, 1)

	tcpConn *tcpsock.TcpConn
	id      uint32
)

func init() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signals
		shutdown <- true
	}()
}

func main() {
	genID()

	proto := &ChatProtocol{}
	proto.OnMessage(onMsg)
	client := tcpsock.NewTcpClient(ServerAddr, proto)
	client.OnConnect(onConnect)
	client.OnClose(onClose)
	go client.Run()
	go input()
	<-shutdown
	client.Close()
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
	log.Printf("%d: %s\n", p.PlayerID, string(p.Body))
}

func genID() {
	fmt.Printf("pls enter your id: ")
	fmt.Scan(&id)
	fmt.Println("your id is:", id)
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
	head.PlayerID = id
	head.BodyLen = uint32(len(s))
	body := make([]byte, head.BodyLen)
	copy(body[:], []byte(s)[:])
	return NewChatPacket(head, body)
}
