package clientsock

import (
	"cacheserver/cfgmgr"
	"cacheserver/errcode"
	"cacheserver/msgnode"
	"errors"
	"fmt"
	"hash/crc32"
	"time"

	. "protocols"

	"tcpsock.v2"
)

type FnWrite = func(b []byte) (n int, err error)
type FnClose = func() error

type client struct {
	conn        *tcpsock.TcpConn
	onWrite     FnWrite
	onClose     FnClose
	recvBuf     []byte
	recvBufLen  int
	bucketChans []chan *msgnode.MsgNode
	msgChan     chan *msgnode.MsgNode
}

func newClient(conn *tcpsock.TcpConn, fnWrite FnWrite, fnClose FnClose, bucketChans []chan *msgnode.MsgNode) *client {
	c := &client{
		conn:        conn,
		onWrite:     fnWrite,
		onClose:     fnClose,
		bucketChans: bucketChans,
		msgChan:     make(chan *msgnode.MsgNode),
	}
	go c.run()
	return c
}

func (self *client) SockHandle() uint64 {
	return self.conn.ID()
}

func (self *client) run() {
	for node := range self.msgChan {
		self.Write(NewPacket(node.Cmd, node.Param, node.Body).Bytes())
	}
}

func (self *client) Read(b []byte) (n int, err error) {
	count := len(b)
	if count+self.recvBufLen > tcpsock.RecvBufLenMax {
		return 0, errors.New("invalid data")
	}

	self.recvBuf = append(self.recvBuf, b[0:count]...)
	self.recvBufLen += count
	offsize := 0
	offset := 0
	var head Head
	for self.recvBufLen-offsize >= SizeOfPacketHead {
		offset = 0
		head.Len = uint32(uint32(self.recvBuf[offsize+3])<<24 | uint32(self.recvBuf[offsize+2])<<16 | uint32(self.recvBuf[offsize+1])<<8 | uint32(self.recvBuf[offsize+0]))
		offset += 4
		head.Cmd = uint8(self.recvBuf[offsize+offset+0])
		offset += 1
		// head.Reserved = uint8(self.recvBuf[offsize+offset+0])
		offset += 1
		head.Param = uint16(uint16(self.recvBuf[offsize+offset+1])<<8 | uint16(self.recvBuf[offsize+offset+0]))
		offset += 2
		pkglen := int(SizeOfPacketHead + head.Len)
		if pkglen >= tcpsock.RecvBufLenMax {
			offsize = self.recvBufLen
			break
		}
		if offsize+pkglen > self.recvBufLen {
			break
		}
		self.process(head.Cmd, head.Param, self.recvBuf[offsize+offset:offsize+offset+int(head.Len)])
		offsize += SizeOfPacketHead
	}

	self.recvBufLen -= offsize
	if self.recvBufLen > 0 {
		self.recvBuf = self.recvBuf[offsize : offsize+self.recvBufLen]
	} else {
		self.recvBuf = nil
	}
	return len(b), nil
}

func (self *client) process(cmd uint8, param uint16, body []byte) {
	self.conn.RawConn().SetReadDeadline(time.Now().Add(time.Duration(cfgmgr.ClientReadDeadline()) * time.Second))
	switch cmd {
	case CM_PING:
		self.Write(NewPacket(SM_PING, 0, nil).Bytes())
	case CM_GET:
		if int(param) >= len(body) {
			self.Write(NewPacket(SM_GET, errcode.ErrInvalidGetReq, nil).Bytes())
			return
		}
		idx := int(crc32.ChecksumIEEE(body[:param])) / len(self.bucketChans)
		self.bucketChans[idx] <- msgnode.New(self.msgChan, cmd, param, body[param:])
	default:
		fmt.Println("?????")
		self.Close()
	}
}

func (self *client) Write(b []byte) (n int, err error) {
	return self.onWrite(b)
}

func (self *client) Close() error {
	return self.onClose()
}
