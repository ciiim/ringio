package main

import (
	"github.com/ciiim/cloudborad/conf"
	"github.com/ciiim/cloudborad/internal/database"
	"github.com/ciiim/cloudborad/router"
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
	mysqlDataSource := v.GetString("database.mysql.datasource")

	if err := database.InitMysql(mysqlDataSource); err != nil {
		panic("Init Mysql Error: " + err.Error())
	}

	s := server.NewServer("group", serverName, ip)

	for _, node := range nodelist {
		s.AddPeer(node["name"], node["addr"])
	}

	if debug {
		s.DebugOn()
	}
	if apiEnable {
		r := router.InitApiServer(s)
		go r.Run(port)
	}

	s.StartServer()
}
