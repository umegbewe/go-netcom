package tcp

import (
	"encoding/binary"
	"net"
	"time"
)

const (
	recvBufferSize = 2048
	pingInterval   = 0 // in seconds, set to 0 to disable pinging
)

type MsgHandler interface {
	HandleMsg(*MsgData)
}

type MsgData struct {
	Data []byte
	Ext  interface{}
}

type MsgPacker interface {
	PackMsg(*MsgData) []byte
	UnpackMsg([]byte) *MsgData
}

type PingHandler interface {
	HandlePing()
}

type CloseHandler interface {
	HandleClose()
}

type Conn struct {
	conn         net.Conn
	send         chan *MsgData
	msgHandler   MsgHandler
	msgPacker    MsgPacker
	pingHandler  PingHandler
	closeHandler CloseHandler
}

func NewConn(conn net.Conn) *Conn {
	c := &Conn{
		conn:         conn,
		send:         make(chan *MsgData, 64),
		msgHandler:   nil,
		msgPacker:    nil,
		pingHandler:  nil,
		closeHandler: nil,
	}

	go c.recvLoop()
	go c.sendLoop()

	return c
}

func (c *Conn) SetMsgHandler(hdlr MsgHandler) {
	c.msgHandler = hdlr
}

func (c *Conn) SetMsgPacker(packer MsgPacker) {
	c.msgPacker = packer
}

func (c *Conn) SetPing(sec uint, hdlr PingHandler) {
	if sec == 0 {
		return
	}
	c.pingHandler = hdlr
	go func() {
		for range time.Tick(time.Duration(sec) * time.Second) {
			if c.pingHandler != nil {
				c.pingHandler.HandlePing()
			}
		}
	}()
}

func (c *Conn) SetCloseHandler(hdlr CloseHandler) {
	c.closeHandler = hdlr
}

func (c *Conn) RawConn() net.Conn {
	return c.conn
}

func (c *Conn) recvLoop() {
	defer c.Close()
	recvBuffer := make([]byte, recvBufferSize)
	for {
		bytesRead, err := c.conn.Read(recvBuffer)
		if err != nil {
			return
		}
		c.handleRecvBuffer(recvBuffer[:bytesRead])
	}
}

func (c *Conn) handleRecvBuffer(recvBuffer []byte) {
	for len(recvBuffer) > 0 {
		if len(recvBuffer) < 4 {
			return
		}
		msgLength := binary.BigEndian.Uint32(recvBuffer[:4])
		if msgLength == 0 {
			if c.pingHandler != nil {
				c.pingHandler.HandlePing()
			}
			recvBuffer = recvBuffer[4:]
			continue
		}
		if len(recvBuffer) < int(msgLength)+4 {
			return
		}
		msg := &MsgData{
			Data: recvBuffer[4 : msgLength+4],
			Ext:  nil,
		}
		if c.msgPacker != nil {
			msg = c.msgPacker.UnpackMsg(recvBuffer[4 : msgLength+4])
		}
		if c.msgHandler != nil {
			c.msgHandler.HandleMsg(msg)
		}
		recvBuffer = recvBuffer[msgLength+4:]
	}
}

func (c *Conn) sendLoop() {
	defer c.Close()
	for msg := range c.send {
		sendBytes := make([]byte, 4)
		if msg != nil {
			data := msg.Data
			if c.msgPacker != nil {
				sendBytes = make([]byte, 4+len(data))
				binary.BigEndian.PutUint32(sendBytes[:4], uint32(len(data)))
				copy(sendBytes[4:], data)
			} else {
				sendBytes = make([]byte, 4)
				binary.BigEndian.PutUint32(sendBytes[:4], 0)
			}
			_, err := c.conn.Write(sendBytes)
			if err != nil {
				return
			}
		}
	}
}
				
func (c *Conn) SendMsg(msg *MsgData) error {
	select {
	case c.send <- msg:
		return nil
	default:
		return net.ErrWriteToConnected
	}
}
				
func (c *Conn) Close() error {
	err := c.conn.Close()
	if c.closeHandler != nil {
		c.closeHandler.HandleClose()
	}
	return err
}



