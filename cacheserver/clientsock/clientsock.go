package clientsock

import (
	"fmt"
	"sync"
	"time"

	"github.com/ecofast/ecocache/cacheserver/cfgmgr"
	"github.com/ecofast/ecocache/cacheserver/msgnode"
	"github.com/ecofast/tcpsock.v2"
)

type listenSock struct {
	*tcpsock.TcpServer
	bucketChans []chan *msgnode.MsgNode
}

var (
	sock *listenSock
)

func Setup() {
	fmt.Printf("public addr: %s\n", cfgmgr.PublicAddr())
}

func Run(exitChan chan struct{}, bucketChans []chan *msgnode.MsgNode, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	sock = &listenSock{}
	sock.TcpServer = tcpsock.NewTcpServer(fmt.Sprintf(":%d", cfgmgr.PublicPort()), sock.onConnect, sock.onDisconnect, nil)
	sock.bucketChans = bucketChans
	sock.Serve()
	<-exitChan
	sock.Close()
}

func (self *listenSock) onConnect(conn *tcpsock.TcpConn) tcpsock.TcpSession {
	conn.RawConn().SetReadDeadline(time.Now().Add(time.Duration(cfgmgr.ClientReadDeadline()) * time.Second))
	return newClient(conn, conn.Write, conn.Close, self.bucketChans)
}

func (self *listenSock) onDisconnect(conn *tcpsock.TcpConn) {
	//
}
