package serversock

import (
	"fmt"
	"net"
	"sync"
	"time"

	"loadbalancer/cfgmgr"

	"tcpsock.v2"
)

type listenSock struct {
	*tcpsock.TcpServer
}

var (
	sock *listenSock
)

func Setup() {
	fmt.Printf("server listen port: %d\n", cfgmgr.ServerListenPort())
}

func Run(exitChan chan struct{}, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	sock = &listenSock{}
	sock.TcpServer = tcpsock.NewTcpServer(fmt.Sprintf(":%d", cfgmgr.ServerListenPort()), sock.onConnect, sock.onDisconnect, sock.onCheckIP)
	sock.Serve()
	<-exitChan
	sock.Close()
}

func (self *listenSock) onConnect(conn *tcpsock.TcpConn) tcpsock.TcpSession {
	conn.RawConn().SetReadDeadline(time.Now().Add(time.Duration(cfgmgr.ServerReadDeadline()) * time.Second))
	return newServer(conn)
}

func (self *listenSock) onDisconnect(conn *tcpsock.TcpConn) {
	//
}

func (self *listenSock) onCheckIP(ip net.Addr) bool {
	return true
}
