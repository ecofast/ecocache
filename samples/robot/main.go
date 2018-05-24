package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	. "github.com/ecofast/rtl/sysutils"
)

var (
	shutdown = make(chan bool, 1)
	robots   []*client
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
	if len(os.Args) != 3 {
		panic("invalid command line params")
	}
	num := StrToInt(os.Args[2])
	robots = make([]*client, num)
	for i := 0; i < num; i++ {
		if i%200 == 0 {
			time.Sleep(200 * time.Millisecond)
		}

		go func(idx int) {
			robots[idx] = newTcpClient(os.Args[1])
			robots[idx].Open()
		}(i)
	}
	<-shutdown
	for i := 0; i < num; i++ {
		robots[i].Close()
	}
}
