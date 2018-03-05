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
</br>
There're more detailed [chatroom](https://github.com/ecofast/tcpsock/tree/master/samples/chatroom) demos available which use custom binary protocols.
