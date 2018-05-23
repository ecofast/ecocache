package clientsock

import (
	"fmt"
	"sync"
	"time"

	"github.com/ecofast/ecocache/loadbalancer/cfgmgr"
	"github.com/ecofast/tcpsock.v2"
)

type listenSock struct {
	*tcpsock.TcpServer
}

var (
	sock *listenSock
)

func Setup() {
	fmt.Printf("client listen port: %d\n", cfgmgr.ClientListenPort())
	fmt.Printf("client read deadline: %d(s)\n", cfgmgr.ClientReadDeadline())
}

func Run(exitChan chan struct{}, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	sock = &listenSock{}
	sock.TcpServer = tcpsock.NewTcpServer(fmt.Sprintf(":%d", cfgmgr.ClientListenPort()), sock.onConnect, sock.onDisconnect, nil)
	sock.Serve()
	<-exitChan
	sock.Close()
}

func (self *listenSock) onConnect(conn *tcpsock.TcpConn) tcpsock.TcpSession {
	conn.RawConn().SetReadDeadline(time.Now().Add(time.Duration(cfgmgr.ClientReadDeadline()) * time.Second))
	return newClient(conn.ID(), conn.RawConn().RemoteAddr(), conn.Write, conn.Close)
}

func (self *listenSock) onDisconnect(conn *tcpsock.TcpConn) {
	//
}
