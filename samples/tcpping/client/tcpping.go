package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	. "github.com/ecofast/rtl/netutils"
	"github.com/ecofast/tcpsock"
	. "github.com/ecofast/tcpsock/samples/tcpping/protocol"
)

type pingStats struct {
	sendNum int
	lags    []int
}

var (
	shutdown = make(chan bool, 1)

	tcpConn *tcpsock.TcpConn
	packet  *PingPacket

	packetLen    int = 32 // byte
	pingInterval int = 1  // second
	pingTimes    int = 10

	canPing  bool = true
	sendTick time.Time

	stats pingStats
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
	parseFlag()
	genPacket()

	client := tcpsock.NewTcpClient(flag.Args()[0], onConnect, onClose, onProtocol)
	go client.Run()

	ticker := time.NewTicker(time.Duration(pingInterval) * time.Second)
	go func() {
		cnt := 0
		for range ticker.C {
			if tcpConn != nil && canPing && stats.sendNum < pingTimes {
				canPing = false
				sendTick = time.Now()
				tcpConn.Write(packet)
				stats.sendNum++
			}
			cnt++
			if cnt > pingTimes {
				shutdown <- true
				break
			}
		}
	}()

	<-shutdown
	ticker.Stop()
	client.Close()
}

func onProtocol() tcpsock.Protocol {
	proto := &PingProtocol{}
	proto.OnMessage(onMsg)
	return proto
}

func onConnect(c *tcpsock.TcpConn) {
	log.Println("successfully connect to server", IPFromNetAddr(c.RawConn().RemoteAddr()))
	tcpConn = c
	fmt.Printf("TCPPing %s with %d bytes of data...\n", flag.Args()[0], packetLen)
}

func onClose(c *tcpsock.TcpConn) {
	printStats()
	log.Println("disconnect from server", IPFromNetAddr(c.RawConn().RemoteAddr()))
	tcpConn = nil
}

func onMsg(c *tcpsock.TcpConn, p *PingPacket) {
	canPing = true
	lag := int(time.Now().Sub(sendTick) / time.Millisecond)
	stats.lags = append(stats.lags, lag)
	fmt.Printf("%d bytes from %s: time=%dms\n", packetLen, IPFromNetAddr(c.RawConn().RemoteAddr()), lag)
}

func parseFlag() {
	flag.IntVar(&packetLen, "p", packetLen, "packet length(byte)")
	flag.IntVar(&pingInterval, "i", pingInterval, "ping interval(second)")
	flag.IntVar(&pingTimes, "t", pingTimes, "Ping times")
	flag.Parse()

	if len(flag.Args()) == 0 {
		fmt.Println("Usage: tcpping ip:port [arguments]\n")
		os.Exit(1)
	}
}

func genPacket() {
	packet = &PingPacket{
		BodyLen: uint16(packetLen) - PacketHeadSize,
		Body:    make([]byte, packetLen),
	}
}

func printStats() {
	fmt.Printf("---%s tcpping statistics---\n", flag.Args()[0])
	fmt.Printf("%d packets transmitted.\n", stats.sendNum)
	if stats.sendNum > 0 {
		sum := 0
		min := stats.lags[0]
		max := 0
		for _, v := range stats.lags {
			sum += v
			if v < min {
				min = v
			}
			if v > max {
				max = v
			}
		}
		fmt.Printf("min/avg/max lag = %d/%d/%d ms\n", min, sum/len(stats.lags), max)
	}
}
