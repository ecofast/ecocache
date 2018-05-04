package main

import (
	"cacheserver/cfgmgr"
	"cacheserver/clientsock"
	"cacheserver/lbsock"
	"cacheserver/lrucache"
	"cacheserver/msgnode"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	shutdown  = make(chan bool, 1)
	exitChan  = make(chan struct{})
	waitGroup = &sync.WaitGroup{}

	bucketChans []chan *msgnode.MsgNode
)

func init() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signals
		shutdown <- true
	}()
}

func main() {
	fmt.Println("ecocache server version 1.0.0, copyright (c) 2017~2018 ecofast")
	setup()
	serve()
}

func setup() {
	cfgmgr.Setup()
	lrucache.Setup()
	lbsock.Setup()
	clientsock.Setup()

	bucketChans = make([]chan *msgnode.MsgNode, cfgmgr.NumLRUCache())
	for i := 0; i < len(bucketChans); i++ {
		bucketChans[i] = make(chan *msgnode.MsgNode)
	}
}

func serve() {
	log.Println("=====cacheserver service start=====")
	waitGroup.Add(3)
	go lrucache.Run(exitChan, bucketChans, waitGroup)
	go lbsock.Run(exitChan, waitGroup)
	go clientsock.Run(exitChan, bucketChans, waitGroup)
	<-shutdown
	close(exitChan)
	waitGroup.Wait()
	log.Println("=====cacheserver service end=====")
}
