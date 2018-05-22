package main

import (
	"bufio"
	"os"
	"os/signal"
	. "protocols"
	"strings"
	"syscall"
)

var (
	shutdown = make(chan bool, 1)

	cli *client
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
	if len(os.Args) == 1 {
		panic("invalid server addr")
	}
	cli = newTcpClient(os.Args[1])
	cli.Open()
	go run()
	<-shutdown
	cli.Close()
}

func run() {
	reader := bufio.NewReader(os.Stdin)
	for {
		if cli == nil {
			break
		}
		if s, err := reader.ReadString('\n'); err == nil {
			s = strings.TrimSpace(s)
			if strings.HasPrefix(s, "get:") {
				idx := strings.Index(s, ":")
				f := s[idx+1:]
				buf := make([]byte, len(f)*2)
				copy(buf, f)
				copy(buf[len(f):], f)
				cli.Write(NewPacket(CM_GET, 0, uint16(len(f)<<8|len(f)), buf).Bytes())
			} else if strings.HasPrefix(s, "set:") {
				idx := strings.Index(s, ":")
				f := s[idx+1:]
				idx = strings.Index(f, "=")
				k := f[:idx]
				v := f[idx+1:]
				buf := make([]byte, len(k)*2+len(v))
				copy(buf, k)
				copy(buf[len(k):], k)
				copy(buf[len(k)*2:], v)
				cli.Write(NewPacket(CM_SET, 0, uint16(uint16(len(k))<<8|uint16(len(k))), buf).Bytes())
			}
			continue
		}
		break
	}
}
