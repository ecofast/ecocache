package serversock

import (
	"errors"
	"fmt"
	"loadbalancer/cfgmgr"
	"loadbalancer/servers"
	"log"
	"time"

	. "protocols"

	"tcpsock.v2"
)

const (
	SizeOfIPv4 = 4
	SizeOfPort = 2
)

type FnWrite = func(b []byte) (n int, err error)
type FnClose = func() error

type server struct {
	conn       *tcpsock.TcpConn
	onWrite    FnWrite
	onClose    FnClose
	recvBuf    []byte
	recvBufLen int
}

func newServer(conn *tcpsock.TcpConn) *server {
	s := &server{
		conn: conn,
	}
	s.onWrite = conn.Write
	s.onClose = conn.Close
	return s
}

func (self *server) SockHandle() uint64 {
	return self.conn.ID()
}

func (self *server) Read(b []byte) (n int, err error) {
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

func (self *server) process(cmd uint8, param uint16, body []byte) {
	self.conn.RawConn().SetReadDeadline(time.Now().Add(time.Duration(cfgmgr.ServerReadDeadline()) * time.Second))

	switch cmd {
	case CM_REGSVR:
		self.regSvr(body)
	case CM_DELSVR:
		self.delSvr()
	case CM_PING:
		self.Write(NewPacket(SM_PING, 0, nil).Bytes())
	default:
		fmt.Println("?????")
		self.Close()
	}
}

func (self *server) Write(b []byte) (n int, err error) {
	if self.onWrite != nil {
		return self.onWrite(b)
	}
	return len(b), nil
}

func (self *server) Close() error {
	self.unreg()
	if self.onClose != nil {
		return self.onClose()
	}
	return nil
}

func (self *server) regSvr(b []byte) {
	if len(b) != SizeOfIPv4+SizeOfPort {
		self.Write(NewPacket(SM_REGSVR, 1, nil).Bytes())
		return
	}
	servers.Store(self.conn.RawConn().RemoteAddr().String(), b)
	self.Write(NewPacket(SM_REGSVR, 0, nil).Bytes())
	// addr := netutils.UInt32ToIPv4(sysutils.BytesToUInt32(b[:4])) + ":" + sysutils.IntToStr(int(sysutils.BytesToUInt16(b[4:])))
	// log.Printf("cache server %s registered", addr)
	log.Printf("num of cache server: %d\n", servers.Count())
}

func (self *server) delSvr() {
	if self.unreg() {
		self.Write(NewPacket(SM_DELSVR, 0, nil).Bytes())
		return
	}
	self.Write(NewPacket(SM_DELSVR, 1, nil).Bytes())
}

func (self *server) unreg() bool {
	if servers.Remove(self.conn.RawConn().RemoteAddr().String()) {
		log.Printf("num of cache server: %d\n", servers.Count())
		return true
	}
	return false
}
