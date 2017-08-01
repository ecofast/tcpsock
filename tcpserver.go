package tcpsock

import (
	"net"
	"sync"
	"sync/atomic"
	"time"

	. "github.com/ecofast/sysutils"
)

type OnTcpConnCallback func(c *TcpConn)

type TcpServer struct {
	listener      *net.TCPListener
	acceptTimeout int
	*tcpSock
	numOfConn     uint32
	waitGroup     *sync.WaitGroup
	onConnConnect OnTcpConnCallback
	onConnClose   OnTcpConnCallback
}

func NewTcpServer(listenPort, acceptTimeout int, protocol Protocol) *TcpServer {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ":"+IntToStr(int(listenPort)))
	CheckError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	CheckError(err)

	return &TcpServer{
		listener:      listener,
		acceptTimeout: acceptTimeout,
		tcpSock: &tcpSock{
			sendBufCap: cSendBufCap,
			recvBufCap: cRecvBufCap,
			proto:      protocol,
			exitChan:   make(chan struct{}),
		},
		numOfConn: 0,
		waitGroup: &sync.WaitGroup{},
	}
}

var (
	connID uint32 // uint64
)

func (self *TcpServer) Serve() {
	self.waitGroup.Add(1)
	defer func() {
		self.listener.Close()
		self.waitGroup.Done()
	}()

	for {
		select {
		case <-self.exitChan:
			return

		default:
		}

		self.listener.SetDeadline(time.Now().Add(time.Duration(self.acceptTimeout) * time.Second))
		conn, err := self.listener.AcceptTCP()
		if err != nil {
			continue
		}

		atomic.AddUint32(&self.numOfConn, 1)
		self.waitGroup.Add(1)
		go func() {
			c := newTcpConn(atomic.AddUint32(&connID, 1), self, conn, self.sendBufCap, self.recvBufCap)
			if self.onConnConnect != nil {
				self.onConnConnect(c)
			}
			c.run()
			self.waitGroup.Done()
		}()
	}
}

func (self *TcpServer) Close() {
	close(self.exitChan)
	self.waitGroup.Wait()
}

func (self *TcpServer) NumOfConn() uint32 {
	return atomic.LoadUint32(&self.numOfConn)
}

func (self *TcpServer) connClose(conn *TcpConn) {
	atomic.AddUint32(&self.numOfConn, ^uint32(0))
	if self.onConnClose != nil {
		self.onConnClose(conn)
	}
}

func (self *TcpServer) OnConnConnect(fn OnTcpConnCallback) {
	self.onConnConnect = fn
}

func (self *TcpServer) OnConnClose(fn OnTcpConnCallback) {
	self.onConnClose = fn
}
