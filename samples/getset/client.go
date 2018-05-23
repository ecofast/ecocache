package main

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/ecofast/ecocache/cacheserver/utils"
	. "github.com/ecofast/ecocache/protocols"
	"github.com/ecofast/tcpsock.v2"
)

type client struct {
	*tcpsock.TcpClient
	recvBuf    []byte
	recvBufLen int
}

func newTcpClient(addr string) *client {
	c := &client{}
	c.TcpClient = tcpsock.NewTcpClient(addr, c.onConnect, c.onDisconnect)
	go c.run()
	return c
}

func (self *client) run() {
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		self.Write(NewPacket(CM_PING, 0, 0, nil).Bytes())
	}
}

func (self *client) SockHandle() uint64 {
	return self.ID()
}

func (self *client) onConnect(c *tcpsock.TcpConn) tcpsock.TcpSession {
	log.Println("successfully connect to server", c.RawConn().RemoteAddr().String())
	return self
}

func (self *client) onDisconnect(c *tcpsock.TcpConn) {
	log.Println("disconnect from server", c.RawConn().RemoteAddr().String())
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
		head.Ret = uint8(self.recvBuf[offsize+offset+0])
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
		self.process(head.Cmd, head.Ret, head.Param, self.recvBuf[offsize+offset:offsize+offset+int(head.Len)])
		offsize += pkglen
	}

	self.recvBufLen -= offsize
	if self.recvBufLen > 0 {
		self.recvBuf = self.recvBuf[offsize : offsize+self.recvBufLen]
	} else {
		self.recvBuf = nil
	}
	return len(b), nil
}

func (self *client) process(cmd, ret uint8, param uint16, body []byte) {
	switch cmd {
	case SM_PING:
		// fmt.Println("SM_PING")
	case SM_GET:
		bucket := utils.BytesToStr(body[:uint8(param>>8)])
		k := utils.BytesToStr(body[uint8(param>>8) : uint8(param>>8)+uint8(param)])
		if ret == 0 {
			v := utils.BytesToStr(body[uint8(param>>8)+uint8(param):])
			fmt.Printf("SM_GET: %s.%s=%s\n", bucket, k, v)
			return
		}
		fmt.Printf("SM_GET: %s.%s=nil\n", bucket, k)
	case SM_SET:
		bucket := utils.BytesToStr(body[:uint8(param>>8)])
		k := utils.BytesToStr(body[uint8(param>>8) : uint8(param>>8)+uint8(param)])
		v := utils.BytesToStr(body[uint8(param>>8)+uint8(param):])
		fmt.Printf("SM_SET: %s.%s=%s", bucket, k, v)
	default:
		fmt.Println("?????")
		self.Close()
	}
}
