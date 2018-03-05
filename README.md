# tcpsock
Package tcpsock provides easy to use interfaces for TCP I/O.</br></br>

# How to use</br>
## server:
```Go
server := tcpsock.NewTcpServer(listenPort, acceptTimeout, onConnConnect, onConnClose, onProtocol)
go server.Serve()
<-shutdown
server.Close()
```
## client:
```Go
client := tcpsock.NewTcpClient(ServerAddr, onConnect, onClose, onProtocol)
go client.Run()
<-shutdown
client.Close()
```
## There're more detailed demos which use custom binary protocols, like:</br>
* [chatroom](https://github.com/ecofast/tcpsock/tree/master/samples/chatroom)</br>
* [tcpping](https://github.com/ecofast/tcpsock/tree/master/samples/tcpping)</br>
