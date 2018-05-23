package cfgmgr

import (
	"log"
	"os"

	"github.com/ecofast/rtl/inifiles"
	"github.com/ecofast/rtl/sysutils"
)

type config struct {
	serverListenPort   int
	serverReadDeadline int // sec
	serverReplicas     int
	clientListenPort   int
	clientReadDeadline int // sec
}

var (
	cfg *config
)

func Setup() {
	iniName := sysutils.ChangeFileExt(os.Args[0], ".ini")
	ini := inifiles.New(iniName, false)
	cfg = &config{
		serverListenPort:   ini.ReadInt("Server", "ListenPort", 24642),
		serverReadDeadline: ini.ReadInt("Server", "ReadDeadline", 5),
		serverReplicas:     ini.ReadInt("Server", "Replicas", 50),
		clientListenPort:   ini.ReadInt("Client", "ListenPort", 12321),
		clientReadDeadline: ini.ReadInt("Client", "ReadDeadline", 2),
	}
	log.Println("configuration has been loaded successfully")
}

func ServerListenPort() int {
	return cfg.serverListenPort
}

func ServerReadDeadline() int {
	return cfg.serverReadDeadline
}

func ServerReplicas() int {
	return cfg.serverReplicas
}

func ClientListenPort() int {
	return cfg.clientListenPort
}

func ClientReadDeadline() int {
	return cfg.clientReadDeadline
}
