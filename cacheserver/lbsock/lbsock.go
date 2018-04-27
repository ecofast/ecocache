package lbsock

import (
	"cacheserver/cfgmgr"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	. "protocols"
	"sync"
	"time"

	"github.com/ecofast/rtl/netutils"
	"tcpsock.v2"
)

type lbSock struct {
	*tcpsock.TcpClient
	recvBuf    []byte
	recvBufLen int
	connected  bool
	reged      bool
	pingCnt    uint8
	lastTick   time.Time
}

var (
	sock *lbSock

	doneChan = make(chan struct{})
	ticker   *time.Ticker
)

func Setup() {
	fmt.Printf("load balancer addr: %s\n", cfgmgr.LoadBalancerAddr())
	fmt.Printf("load balancer reconnect interval: %d\n", cfgmgr.LoadBalancerReConnIntv())
	fmt.Printf("load balancer ping interval: %d\n", cfgmgr.LoadBalancerPingIntv())
}

func Run(exitChan chan struct{}, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	sock = &lbSock{}
	sock.TcpClient = tcpsock.NewTcpClient(cfgmgr.LoadBalancerAddr(), sock.onConnect, sock.onDisconnect)
	sock.Open()
	sock.regSelf()
	ticker = time.NewTicker(time.Second)
	sock.checkConnect()
	<-exitChan
	ticker.Stop()
	if sock.connected && sock.reged {
		sock.unregSelf()
		select {
		case <-doneChan:
		case <-time.After(time.Second):
		}
	}
	sock.Close()
}

func (self *lbSock) SockHandle() uint64 {
	return self.ID()
}

func (self *lbSock) onConnect(c *tcpsock.TcpConn) tcpsock.TcpSession {
	self.connected = true
	self.pingCnt = 0
	self.lastTick = time.Now()
	log.Printf("successfully connected to loadbalancer(%s)\n", c.RawConn().RemoteAddr().String())
	return self
}

func (self *lbSock) onDisconnect(c *tcpsock.TcpConn) {
	self.connected = false
	self.pingCnt = 0
	log.Printf("disconnected from loadbalancer(%s)\n", c.RawConn().RemoteAddr().String())
}

// to do: GetTickCount
//
func (self *lbSock) checkConnect() {
	go func() {
		for tick := range ticker.C {
			if !self.connected {
				if tick.Sub(self.lastTick) >= time.Duration(cfgmgr.LoadBalancerReConnIntv())*time.Second {
					self.lastTick = tick
					self.Open()
					sock.regSelf()
				}
			} else {
				if tick.Sub(self.lastTick) >= time.Duration(cfgmgr.LoadBalancerPingIntv())*time.Second {
					self.lastTick = tick
					self.sendPing()
					self.pingCnt++
					if self.pingCnt >= 5 {
						log.Println("no response from load balancer")
						self.connected = false
						self.Close()
					}
				}
			}
		}
	}()
}

func (self *lbSock) Read(b []byte) (n int, err error) {
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

func (self *lbSock) process(cmd uint8, param uint16, body []byte) {
	switch cmd {
	case SM_REGSVR:
		if param == 0 {
			self.reged = true
			log.Println("successfully registered to load balancer")
		} else {
			log.Println("register to load balancer failed")
		}
	case SM_DELSVR:
		if param == 0 {
			log.Println("successfully unregistered from load balancer")
		} else {
			log.Println("unregister from load balancer failed")
		}
		doneChan <- struct{}{}
	case SM_PING:
		self.pingCnt = 0
	default:
		fmt.Println("?????")
		self.Close()
	}
}

func (self *lbSock) regSelf() {
	bs := make([]byte, 6)
	binary.LittleEndian.PutUint32(bs, netutils.IPv4ToUInt32(cfgmgr.PublicIP()))
	binary.LittleEndian.PutUint16(bs[4:], uint16(cfgmgr.PublicPort()))
	self.send(CM_REGSVR, 0, bs)
}

func (self *lbSock) unregSelf() {
	self.send(CM_DELSVR, 0, nil)
}

func (self *lbSock) sendPing() {
	self.send(CM_PING, 0, nil)
}

func (self *lbSock) send(cmd uint8, param uint16, body []byte) {
	if self.connected {
		self.Write(NewPacket(cmd, param, body).Bytes())
	}
}
