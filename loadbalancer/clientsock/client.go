package clientsock

import (
	"errors"
	"fmt"
	"loadbalancer/servers"
	"net"

	. "protocols"

	"tcpsock.v2"
)

type FnWrite = func(b []byte) (n int, err error)
type FnClose = func() error

type client struct {
	sockHandle uint64
	remoteAddr net.Addr
	onWrite    FnWrite
	onClose    FnClose
	recvBuf    []byte
	recvBufLen int
}

func newClient(handle uint64, remoteAddr net.Addr, fnWrite FnWrite, fnClose FnClose) *client {
	return &client{
		sockHandle: handle,
		remoteAddr: remoteAddr,
		onWrite:    fnWrite,
		onClose:    fnClose,
	}
}

func (self *client) SockHandle() uint64 {
	return self.sockHandle
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
	for self.recvBufLen-offsize > SizeOfPacketHead {
		offset = 0
		head.Len = uint32(uint32(self.recvBuf[offsize+3])<<24 | uint32(self.recvBuf[offsize+2])<<16 | uint32(self.recvBuf[offsize+1])<<8 | uint32(self.recvBuf[offsize+0]))
		if head.Len != 0 {
			self.Close()
			return 0, errors.New("invalid request")
		}
		offset += SizeOfPacketHeadLen
		head.Cmd = uint8(self.recvBuf[offsize+offset+0])
		self.process(head.Cmd)
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

func (self *client) process(cmd uint8) {
	switch cmd {
	case CM_REQSVR:
		if v := servers.Get(self.remoteAddr.String()); v != nil {
			self.Write(NewPacket(SM_REQSVR, 0, v).Bytes())
			return
		}
		self.Write(NewPacket(SM_REQSVR, 1, nil).Bytes())
	default:
		fmt.Println("?????")
		self.Close()
	}
}

func (self *client) Write(b []byte) (n int, err error) {
	if self.onWrite != nil {
		return self.onWrite(b)
	}
	return len(b), nil
}

func (self *client) Close() error {
	if self.onClose != nil {
		return self.onClose()
	}
	return nil
}
