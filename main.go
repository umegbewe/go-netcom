package main

import (
	"fmt"
	"net"

	"github.com/umegbewe/go-netcom/tcp"
)

func main() {

	conn, err := net.Dial("tcp", "localhost:1234")
	if err != nil {
		panic(err)
	}

	tcpConn := tcp.NewConn(conn)

	tcpConn.SetMsgHandler(&MyMsgHandler{})


	tcpConn.SetMsgPacker(&MyMsgPacker{})

	tcpConn.SetPing(30, &MyPingHandler{})

	tcpConn.SetCloseHandler(&MyCloseHandler{})

	msg := &tcp.MsgData{
		Data: []byte("Hello, world!"),
		Ext:  nil,
	}
	err = tcpConn.SendMsg(msg)
	if err != nil {
		fmt.Println("Failed to send message:", err)
	}

	fmt.Println("Press ENTER to close the connection...")
	fmt.Scanln()

	err = tcpConn.Close()
	if err != nil {
		fmt.Println("Failed to close connection:", err)
	}
}


type MyMsgHandler struct{}

func (h *MyMsgHandler) HandleMsg(msg *tcp.MsgData) {
	fmt.Println("Received message:", string(msg.Data))
}


type MyMsgPacker struct{}

func (p *MyMsgPacker) PackMsg(msg *tcp.MsgData) []byte {
	return []byte("This is a packed message")
}

func (p *MyMsgPacker) UnpackMsg(data []byte) *tcp.MsgData {
	return &tcp.MsgData{
		Data: []byte("This is an unpacked message"),
		Ext:  nil,
	}
}


type MyPingHandler struct{}

func (h *MyPingHandler) HandlePing() {
	fmt.Println("Sending ping message...")
}

type MyCloseHandler struct{}

func (h *MyCloseHandler) HandleClose() {
	fmt.Println("Connection closed")
}
