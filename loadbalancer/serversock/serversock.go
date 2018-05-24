package serversock

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/ecofast/ecocache/loadbalancer/cfgmgr"
	"github.com/ecofast/ecocache/loadbalancer/servers"
	"github.com/ecofast/tcpsock.v2"
)

type listenSock struct {
	*tcpsock.TcpServer
}

var (
	sock *listenSock
)

func Setup() {
	fmt.Printf("server listen port: %d\n", cfgmgr.ServerListenPort())
	fmt.Printf("server read deadline: %d(s)\n", cfgmgr.ServerReadDeadline())
	fmt.Printf("server replicas: %d\n", cfgmgr.ServerReplicas())
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
	if servers.Remove(conn.RawConn().RemoteAddr().String()) {
		log.Printf("num of cache server: %d\n", servers.Count())
	}
}

func (self *listenSock) onCheckIP(ip net.Addr) bool {
	return true
}
