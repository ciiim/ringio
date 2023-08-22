package server

import (
	"log"
	"net"

	"github.com/ciiim/cloudborad/conf"

	"github.com/ciiim/cloudborad/internal/fs"
)

type Server struct {
	Group *fs.Group
}

func StartLocalServer() {
	v := conf.InitConfig()
	serverName := v.Get("basic.server_name").(string)
	localServer := NewServer("default", serverName, "127.0.0.1")

	localServer.StartServer()
}

/*
ffs is the front file system

it must be a tree structure
*/
func NewServer(groupName, serverName, addr string) *Server {
	if addr == "" {
		addr = GetIP()
	}
	fs.DebugOn()
	log.Println("[Server] New server", serverName, addr)
	ffs := fs.NewDTFS(*fs.NewDPeer("front0_"+serverName+"_"+groupName, addr+":"+fs.FRONT_PORT, 20, nil), "./front0_"+serverName+"_"+groupName)
	sfs := fs.NewDFS(*fs.NewDPeer("store0_"+serverName+"_"+groupName, addr+":"+fs.FILE_STORE_PORT, 20, nil), "./store0_"+serverName+"_"+groupName, 1024*1024*1024, nil)
	if ffs == nil || sfs == nil {
		log.Fatal("New server failed")
	}
	server := &Server{
		Group: fs.NewGroup(groupName, ffs),
	}
	server.Group.UseFS(sfs)
	return server
}

func (s *Server) StartServer() {
	r := initRoute(s)

	go s.Group.Serve()
	r.Run(":8080")
}

func (s *Server) Join(peerName, peerAddr string) error {
	err := s.Group.Join(peerName, peerAddr)
	if err != nil {
		return err
	}
	log.Println("[Server] Join cluster success")
	return nil
}

func (s *Server) Quit() {
	s.Group.Quit()
}

func (s *Server) Close() error {
	s.Quit()
	return s.Group.Close()
}

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
