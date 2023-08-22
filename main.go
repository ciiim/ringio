package main

import (
	"log"

	"github.com/ciiim/cloudborad/conf"
	"github.com/ciiim/cloudborad/server"
)

func main() {
	ip := server.GetIP()
	v := conf.InitConfig()
	serverName := v.Get("basic.server_name").(string)
	server := server.NewServer("test_server", serverName, ip)
	log.Println("Server IP:", ip)
	server.StartServer()
}
