package main

import (
	"log"
	"time"

	"github.com/ciiim/cloudborad/conf"
	"github.com/ciiim/cloudborad/internal/database"
	"github.com/ciiim/cloudborad/router"
	"github.com/ciiim/cloudborad/server"
	"github.com/ciiim/cloudborad/service"
)

func main() {

	v := conf.InitConfig()

	//mysql

	mysqlDataSource := v.GetString("database.mysql.datasource")

	if err := database.InitMysql(mysqlDataSource); err != nil {
		log.Fatalln("Init Mysql Error: " + err.Error())
	}
	log.Println("[Init] mysql is connected.")

	//redis

	redisEnable := v.GetBool("database.redis.enable")
	if redisEnable {
		redisHost := v.GetString("database.redis.host")
		redisPort := v.GetString("database.redis.port")
		redisPassword := v.GetString("database.redis.password")
		if err := database.InitRedis(redisHost+":"+redisPort, redisPassword, 1); err != nil {
			log.Fatalln("Init Redis Error: " + err.Error())
		}
		log.Println("[Init] redis is connected.")
	}

	//file server

	nodelist := conf.GetNodes(v)
	serverName := v.GetString("server.file_server_name")
	serverPort := v.GetString("server.file_server_port")

	s := server.NewServer("group", serverName, server.GetIP(), serverPort)

	for _, node := range nodelist {
		s.AddPeer(node["name"], node["addr"])
	}

	//debug

	debug := v.GetBool("debug")

	if debug {
		s.DebugOn()
	}

	//api server

	apiEnable := v.GetBool("server.api_server_enable")

	if apiEnable {
		log.Println("[Init] api server enable.")

		port := v.GetString("server.api_server_port")

		smtpHost := v.GetString("smtp.host")
		smtpPort := v.GetString("smtp.port")
		email := v.GetString("smtp.email")
		password := v.GetString("smtp.password")
		if smtpHost == "" || smtpPort == "" || email == "" || password == "" {
			panic("smtp config error")
		}
		serv := service.NewService(s)
		serv.SetEmailConfig(service.EmailConfig{
			Smtp: &service.SmtpConfig{
				Host:     smtpHost,
				Port:     smtpPort,
				Email:    email,
				Password: password,
			},
			VerifyCodeLen:        6,
			VerifyCodeExpireTime: 2 * time.Minute,
		})

		r := router.InitApiServer(serv)
		go r.Run(port)
	}
	log.Println("[Init] server is starting...")
	s.StartServer()
}
