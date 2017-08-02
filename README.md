# tcpsock
Package tcpsock provides easy to use interfaces for TCP I/O.</br></br>

# How to use
proto := &YourCustomProtocol{}</br>
server := tcpsock.NewTcpServer(listenPort, acceptTimeout, proto)</br>
go server.Serve()</br>
<-shutdown</br>
server.Close()</br></br>

There's a more detailed [chatroom](https://github.com/ecofast/tcpsock/tree/master/samples/chatroom) demo available which uses a custom binary protocol.
