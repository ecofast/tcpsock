# tcpsock
Package tcpsock provides easy to use interfaces for TCP I/O.</br></br>

# How to use</br>
## server:
server := tcpsock.NewTcpServer(listenPort, acceptTimeout, onConnConnect, onConnClose, onProtocol)</br>
go server.Serve()</br>
<-shutdown</br>
server.Close()</br></br>
## client:
client := tcpsock.NewTcpClient(ServerAddr, onConnect, onClose, onProtocol)</br>
go client.Run()</br>
<-shutdown</br>
client.Close()</br></br>

There's a more detailed [chatroom](https://github.com/ecofast/tcpsock/tree/master/samples/chatroom) demo available which uses a custom binary protocol.
