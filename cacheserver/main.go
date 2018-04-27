package main

import (
	"cacheserver/cfgmgr"
	"cacheserver/clientsock"
	"cacheserver/lbsock"
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
	lbsock.Setup()
	clientsock.Setup()
}

func serve() {
	log.Println("=====cacheserver service start=====")
	waitGroup.Add(2)
	go lbsock.Run(exitChan, waitGroup)
	go clientsock.Run(exitChan, waitGroup)
	<-shutdown
	close(exitChan)
	waitGroup.Wait()
	log.Println("=====cacheserver service end=====")
}
