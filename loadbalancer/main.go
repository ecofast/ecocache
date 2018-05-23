package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/ecofast/ecocache/loadbalancer/cfgmgr"
	"github.com/ecofast/ecocache/loadbalancer/clientsock"
	"github.com/ecofast/ecocache/loadbalancer/servers"
	"github.com/ecofast/ecocache/loadbalancer/serversock"
)

var (
	shutdown  = make(chan bool, 1)
	exitChan  = make(chan struct{})
	waitGroup = &sync.WaitGroup{}
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
	fmt.Println("ecocache loadbalancer version 1.0.0, copyright (c) 2017~2018 ecofast")

	setup()
	serve()
}

func setup() {
	cfgmgr.Setup()
	servers.Setup()
	serversock.Setup()
	clientsock.Setup()
}

func serve() {
	log.Println("=====loadbalancer service start=====")
	waitGroup.Add(2)
	go serversock.Run(exitChan, waitGroup)
	go clientsock.Run(exitChan, waitGroup)
	<-shutdown
	close(exitChan)
	waitGroup.Wait()
	log.Println("=====loadbalancer service end=====")
}
