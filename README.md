# tcpsock
Package tcpsock provides easy to use interfaces for TCP I/O.</br></br>

# How to use</br>
## server:
proto := &YourCustomProtocol{}</br>
proto.OnMessage(onMsg)</br>
server := tcpsock.NewTcpServer(listenPort, acceptTimeout, proto)</br>
server.OnConnConnect(onConnConnect)</br>
server.OnConnClose(onConnClose)</br>
go server.Serve()</br>
<-shutdown</br>
server.Close()</br></br>
## client:
proto := &YourCustomProtocol{}</br>
proto.OnMessage(onMsg)</br>
client := tcpsock.NewTcpClient(ServerAddr, proto)</br>
client.OnConnect(onConnect)</br>
client.OnClose(onClose)</br>
go client.Run()</br>
<-shutdown</br>
client.Close()</br></br>

There's a more detailed [chatroom](https://github.com/ecofast/tcpsock/tree/master/samples/chatroom) demo available which uses a custom binary protocol.
