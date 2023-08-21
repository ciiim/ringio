package server

import (
	"errors"

	"github.com/gin-gonic/gin"
)

func initRoute(s *Server) *gin.Engine {
	r := gin.Default()
	apiGroup := r.Group("/api/v1")
	{
		apiGroup.GET("/space/:key/:path", s.GetDir)
		apiGroup.PUT("/space/:key", s.NewBoard)

		apiGroup.GET("/cluster", s.GetCluster)
		apiGroup.PUT("/cluster/:name/:addr", s.JoinCluster)
		apiGroup.DELETE("/cluster", s.QuitCluster)
	}
	adminGroup := r.Group("/admin")
	{
		adminGroup.GET("/index")
	}
	return r
}

func (s *Server) GetCluster(ctx *gin.Context) {
	list := s.Group.FrontSystem.Peer().PList()
	ctx.JSON(200, gin.H{
		"errcode":  0,
		"success":  true,
		"peernum":  len(list),
		"peerlist": list.Readable(),
	})
}

func (s *Server) QuitCluster(ctx *gin.Context) {
	s.Quit()
	ctx.JSON(200, gin.H{
		"errcode": 0,
		"success": true,
	})
}

func (s *Server) JoinCluster(ctx *gin.Context) {
	err := s.Join(ctx.Param("name"), ctx.Param("addr"))
	ctx.JSON(200, gin.H{
		"msg":     err,
		"success": errors.Is(err, nil),
	})
}

/*
Space API
*/

func (s *Server) NewBoard(ctx *gin.Context) {
	err := s.Group.FrontSystem.Store(ctx.Param("key"), ctx.Param("key"), nil)
	ctx.JSON(200, gin.H{
		"msg":     err,
		"success": errors.Is(err, nil),
	})
}

func (s *Server) MkDir(ctx *gin.Context) {
	err := s.Group.FrontSystem.Store(ctx.Param("key"), ctx.Param("path"), nil)
	ctx.JSON(200, gin.H{
		"msg":     err,
		"success": errors.Is(err, nil),
	})
}

func (s *Server) GetDir(ctx *gin.Context) {
	file, err := s.Group.FrontSystem.Get(ctx.Param("key") + ctx.Param("path"))
	ctx.JSON(200, gin.H{
		"msg":     err,
		"success": errors.Is(err, nil),
		"dir":     file.Stat().SubDir(),
	})

}
