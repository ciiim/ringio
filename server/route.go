package server

import (
	"errors"
	"net/http"

	"github.com/ciiim/cloudborad/internal/fs"
	"github.com/gin-gonic/gin"
)

func initRoute(s *Server) *gin.Engine {
	if fs.IsDebug() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	apiGroup := r.Group("/api/v1")
	{
		apiGroup.GET("/board", s.GetDir)
		apiGroup.PUT("/board", s.MkDir)
		apiGroup.PUT("/board/:key", s.NewBoard)

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
	dpeerList := make([]fs.DPeerInfo, 0, len(list))
	for _, peer := range list {
		dpeerList = append(dpeerList, peer.(fs.DPeerInfo))
	}
	ctx.JSON(http.StatusOK, gin.H{
		"meg":      "success",
		"success":  true,
		"peernum":  len(list),
		"peerlist": dpeerList,
	})
}

func (s *Server) QuitCluster(ctx *gin.Context) {
	s.Quit()
	ctx.JSON(http.StatusOK, gin.H{
		"msg":     "success",
		"success": true,
	})
}

func (s *Server) JoinCluster(ctx *gin.Context) {
	err := s.Join(ctx.Param("name"), ctx.Param("addr"))
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"msg":     err.Error(),
			"success": errors.Is(err, nil),
		})
		return
	} else {
		ctx.JSON(http.StatusOK, gin.H{
			"msg":     "success",
			"success": errors.Is(err, nil),
		})
	}
}

/*
Space API
*/

func (s *Server) NewBoard(ctx *gin.Context) {
	err := s.Group.NewBorad(ctx.Param("key"))
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"msg":     err.Error(),
			"success": errors.Is(err, nil),
		})
		return
	} else {
		ctx.JSON(http.StatusOK, gin.H{
			"msg":     "success",
			"success": errors.Is(err, nil),
		})
	}

}

func (s *Server) MkDir(ctx *gin.Context) {
	key, _ := ctx.GetQuery("key")
	base, _ := ctx.GetQuery("base")
	if base == "root" || base == "" {
		base = "."
	}
	dirName, _ := ctx.GetQuery("dir")
	err := s.Group.Mkdir(key, base, dirName)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"msg":     err.Error(),
			"success": errors.Is(err, nil),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"msg":     "success",
		"success": errors.Is(err, nil),
	})
}

func (s *Server) GetDir(ctx *gin.Context) {
	key, _ := ctx.GetQuery("key")
	base, _ := ctx.GetQuery("base")
	if base == "root" {
		base = "."
	}
	dirName, _ := ctx.GetQuery("dir")
	file, err := s.Group.GetDir(key, base, dirName)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"msg":     err.Error(),
			"success": errors.Is(err, nil),
			"sub":     nil,
		})
	} else {
		info := file.Stat().SubDir()
		ctx.JSON(http.StatusOK, gin.H{
			"msg":     "success",
			"success": errors.Is(err, nil),
			"sub":     info,
		})
	}

}
