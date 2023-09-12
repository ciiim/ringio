package server

import (
	"fmt"
	"testing"
)

func TestServer(t *testing.T) {
	var servers []*Server
	port := 9630
	for i := 0; i < 10; i++ {
		servers = append(servers, NewServer(fmt.Sprintf("node-%d", i), "", fmt.Sprintf("%d", port), nil, OPTION_NO_FRONT, OPTION_NO_STORE))
		port++
	}
	for i := 0; i < 9; i++ {
		go servers[i].StartServer()
	}
	servers[len(servers)-1].StartServer()
}
