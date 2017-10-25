package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	. "github.com/ecofast/rtl/netutils"

	"github.com/ecofast/tcpsock"

	. "github.com/ecofast/tcpsock/samples/chatroom/protocol"
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

func main() {
	clients = make(map[uint32]*tcpsock.TcpConn)

	server := tcpsock.NewTcpServer(9999, 5, onConnConnect, onConnClose, onProtocol)
	log.Println("=====service start=====")
	go server.Serve()

	ticker := time.NewTicker(2 * time.Minute)
	go func() {
		for range ticker.C {
			//log.Printf("num of conn: %d\n", server.NumOfConn())
			// broadcast()
		}
	}()

	<-shutdown
	log.Println("server shutdown...")
	server.Close()
	log.Println("=====service end=====")
}

func onConnConnect(conn *tcpsock.TcpConn) {
	conn.Write(genChatPacket("master01", fmt.Sprintf("Welcome! Your IP is %s", IPFromNetAddr(conn.RawConn().RemoteAddr()))))

	mutex.Lock()
	defer mutex.Unlock()
	clients[conn.ID()] = conn
}

func onConnClose(conn *tcpsock.TcpConn) {
	mutex.Lock()
	defer mutex.Unlock()
	delete(clients, conn.ID())
}

func onMsg(conn *tcpsock.TcpConn, p *ChatPacket) {
	fmt.Println(p)

	mutex.Lock()
	defer mutex.Unlock()
	for _, c := range clients {
		c.Write(p)
	}
}

func onProtocol() tcpsock.Protocol {
	proto := &ChatProtocol{}
	proto.OnMessage(onMsg)
	return proto
}

func genChatPacket(userName, words string) *ChatPacket {
	var head PacketHead
	head.Signature = ChatSignature
	copy(head.UserName[:], ([]byte(userName[:]))[:])
	body := []byte(words)
	head.BodyLen = uint32(len(body))
	return NewChatPacket(head, body)
}

func broadcast() {
	mutex.Lock()
	defer mutex.Unlock()
	packet := genChatPacket("master01", "broadcast test")
	for _, c := range clients {
		c.Write(packet)
	}
}
