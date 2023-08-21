package main

import "github.com/ciiim/cloudborad/server"

func main() {
	server := server.NewServer("test_server", "server0", "127.0.0.1")
	server.StartServer()
}
