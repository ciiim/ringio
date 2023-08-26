package server

import (
	"log"
	"net"
)

func GetIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Println(err)
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
