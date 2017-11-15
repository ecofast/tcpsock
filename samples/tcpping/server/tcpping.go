package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	. "github.com/ecofast/rtl/netutils"
	"github.com/ecofast/tcpsock"
	. "github.com/ecofast/tcpsock/samples/tcpping/protocol"
)

var (
	shutdown = make(chan bool, 1)

	listenPort int = 12345
)

func init() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signals
		shutdown <- true
	}()
}

func parseFlag() {
	flag.IntVar(&listenPort, "p", listenPort, "listen port")
	flag.Parse()
}

func main() {
	parseFlag()

	fmt.Printf("tcpping listening on port: %d\n", listenPort)
	server := tcpsock.NewTcpServer(listenPort, 2, onConnConnect, onConnClose, onProtocol)
	log.Println("=====service start=====")
	go server.Serve()

	<-shutdown
	log.Println("shutdown server")
	server.Close()
	log.Println("=====service stop=====")
}

func onConnConnect(conn *tcpsock.TcpConn) {
	log.Printf("accept connection from %s\n", IPFromNetAddr(conn.RawConn().RemoteAddr()))
}

func onConnClose(conn *tcpsock.TcpConn) {
	log.Printf("connection closed from %s\n", IPFromNetAddr(conn.RawConn().RemoteAddr()))
}

func onMsg(conn *tcpsock.TcpConn, p *PingPacket) {
	log.Printf("recved ping message from %s with %d bytes of data\n", IPFromNetAddr(conn.RawConn().RemoteAddr()), PacketHeadSize+p.BodyLen)
	conn.Write(p)
}

func onProtocol() tcpsock.Protocol {
	proto := &PingProtocol{}
	proto.OnMessage(onMsg)
	return proto
}
