package main

import (
	"github.com/ciiim/cloudborad/conf"
	"github.com/ciiim/cloudborad/server"
)

func main() {
	ip := server.GetIP()
	v := conf.InitConfig()
	serverName := v.GetString("server.server_name")
	debug := v.GetBool("debug")
	port := v.GetString("server.server_port")
	nodelist := conf.GetNodes(v)
	s := server.NewServer("test_server", serverName, ip)
	for _, node := range nodelist {
		s.AddPeer(node["name"], node["addr"])
	}
	if debug {
		s.DebugOn()
	}
	s.StartServer(port)
}
