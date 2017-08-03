package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"tcpsock"
	. "tcpsock/samples/chatroom/protocol"
	"time"

	. "github.com/ecofast/sysutils"
)

var (
	shutdown = make(chan bool, 1)

	mutex   sync.Mutex
	clients map[uint32]*tcpsock.TcpConn
)

func init() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signals
		shutdown <- true
	}()
}

func onConnConnect(conn *tcpsock.TcpConn) {
	// conn.Send(genChatPacket())
	mutex.Lock()
	defer mutex.Unlock()
	clients[conn.ID()] = conn
}

func onConnClose(conn *tcpsock.TcpConn) {
	mutex.Lock()
	defer mutex.Unlock()
	delete(clients, conn.ID())
}

func genChatPacket() *ChatPacket {
	var head PacketHead
	head.Signature = ChatSignature
	head.PlayerID = 555555555
	s := "current time is " + TimeToStr(time.Now())
	head.BodyLen = uint32(len(s))
	body := make([]byte, int(head.BodyLen))
	copy(body[:], []byte(s)[:])
	return NewChatPacket(head, body)
}

func broadcast() {
	mutex.Lock()
	defer mutex.Unlock()
	packet := genChatPacket()
	for _, c := range clients {
		c.Write(packet)
	}
}

func onMsg(conn *tcpsock.TcpConn, p *ChatPacket) {
	mutex.Lock()
	defer mutex.Unlock()
	for _, c := range clients {
		c.Write(p)
	}
}

func main() {
	clients = make(map[uint32]*tcpsock.TcpConn)

	proto := &ChatProtocol{}
	proto.OnMessage(onMsg)
	server := tcpsock.NewTcpServer(9999, 2, proto)
	server.OnConnConnect(onConnConnect)
	server.OnConnClose(onConnClose)
	log.Println("=====service start=====")
	go server.Serve()

	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for range ticker.C {
			log.Printf("num of conn: %d\n", server.NumOfConn())
			broadcast()
		}
	}()

	<-shutdown
	server.Close()
	log.Println("=====service end=====")
}
