package main

import (
	"github.com/ciiim/cloudborad/conf"
	"github.com/ciiim/cloudborad/server"
)

func main() {
	ip := server.GetIP()
	v := conf.InitConfig()

	debug := v.GetBool("debug")

	serverName := v.GetString("server.file_server_name")

	port := v.GetString("server.api_server_port")
	apiEnable := v.GetBool("server.api_server_enable")

	nodelist := conf.GetNodes(v)

	s := server.NewServer("group", serverName, ip)

	for _, node := range nodelist {
		s.AddPeer(node["name"], node["addr"])
	}

	if debug {
		s.DebugOn()
	}

	s.StartServer(port, apiEnable)
}
