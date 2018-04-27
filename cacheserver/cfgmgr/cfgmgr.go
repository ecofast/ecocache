package cfgmgr

import (
	"fmt"
	"log"
	"os"

	"github.com/ecofast/rtl/inifiles"
	"github.com/ecofast/rtl/sysutils"
)

type config struct {
	publicIP               string
	publicPort             int
	loadBalancerIP         string
	loadBalancerPort       int
	loadBalancerReConnIntv int
	loadBalancerPingIntv   int
	clientReadDeadline     int
}

var (
	cfg *config
)

func Setup() {
	iniName := sysutils.ChangeFileExt(os.Args[0], ".ini")
	ini := inifiles.New(iniName, false)
	cfg = &config{
		publicIP:               ini.ReadString("Setup", "PublicIP", ""),
		publicPort:             ini.ReadInt("Setup", "PublicPort", 0),
		loadBalancerIP:         ini.ReadString("LoadBalancer", "IP", ""),
		loadBalancerPort:       ini.ReadInt("LoadBalancer", "Port", 0),
		loadBalancerReConnIntv: ini.ReadInt("LoadBalancer", "ReConnIntv", 10),
		loadBalancerPingIntv:   ini.ReadInt("LoadBalancer", "PingIntv", 5),
		clientReadDeadline:     ini.ReadInt("Client", "ReadDeadline", 15),
	}
	if isValidIP(cfg.publicIP) && isValidPort(cfg.publicPort) && isValidIP(cfg.loadBalancerIP) && isValidPort(cfg.loadBalancerPort) {
		log.Println("configuration has been loaded successfully")
		return
	}
	panic("invalid configuration!")
}

func isValidIP(ip string) bool {
	return ip != ""
}

func isValidPort(port int) bool {
	return port > 1024 && port < 65536
}

func PublicIP() string {
	return cfg.publicIP
}

func PublicPort() uint16 {
	return uint16(cfg.publicPort)
}

func LoadBalancerAddr() string {
	return fmt.Sprintf("%s:%d", cfg.loadBalancerIP, cfg.loadBalancerPort)
}

func LoadBalancerReConnIntv() int {
	return cfg.loadBalancerReConnIntv
}

func LoadBalancerPingIntv() int {
	return cfg.loadBalancerPingIntv
}

func ClientReadDeadline() int {
	return cfg.clientReadDeadline
}
