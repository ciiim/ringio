package router

import (
	"errors"
	"net/http"

	"github.com/ciiim/cloudborad/internal/fs"
	"github.com/ciiim/cloudborad/server"
	"github.com/gin-gonic/gin"
)

func debugApi(r *gin.Engine) {
	debugGroup := r.Group("/api/debug")
	{
		debugGroup.GET("/getblock")
		debugGroup.PUT("/storeblock")
	}
}

type ApiServer struct {
	r          *gin.Engine
	fileServer *server.Server
}

func InitApiServer(fileServer *server.Server) *ApiServer {
	if fs.IsDebug() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	if fs.IsDebug() {
		debugApi(r)
	}
	as := &ApiServer{
		r:          r,
		fileServer: fileServer,
	}

	APIGroup := r.Group("/api/v1")
	{
		fsAPIGroup := APIGroup.Group("fs")
		{
			fsAPIGroup.GET("/board", as.GetDir)
			fsAPIGroup.PUT("/board", as.MkDir)
			fsAPIGroup.PUT("/board/:key", as.NewBoard)

			fsAPIGroup.GET("/cluster", as.GetCluster)
			fsAPIGroup.POST("/cluster", as.JoinCluster)
			fsAPIGroup.DELETE("/cluster", as.QuitCluster)
		}

		authAPIGroup := APIGroup.Group("auth")
		{
			//登录
			authAPIGroup.POST("/login")

			//注册
			authAPIGroup.POST("/register")

			//登出
			authAPIGroup.POST("/logout")

			//使用邮箱激活验证账号
			authAPIGroup.POST("/verify")
		}

		adminGroup := as.r.Group("/internal/admin", adminAuth())
		{
			as.r.LoadHTMLGlob("server/admin/*")

			//简易节点操作
			adminGroup.GET("/peer", as.AdminPage)

			//关闭节点并退出集群
			adminGroup.POST("/shutdown", as.Shutdown)
		}
	}

	return as
}

func (a *ApiServer) GetCluster(ctx *gin.Context) {
	list := a.fileServer.Group.FrontSystem.Peer().PList()
	dpeerList := fs.PeerInfoListToDpeerInfoList(a.fileServer.Group.FrontSystem.Peer().PList())
	ctx.JSON(http.StatusOK, gin.H{
		"meg":      "success",
		"success":  true,
		"peernum":  len(list),
		"peerlist": dpeerList,
	})
}

func (a *ApiServer) QuitCluster(ctx *gin.Context) {
	a.fileServer.Quit()
	ctx.JSON(http.StatusOK, gin.H{
		"msg":     "success",
		"success": true,
	})
}

func (a *ApiServer) JoinCluster(ctx *gin.Context) {
	peerName := ctx.PostForm("peerName")
	peerAddr := ctx.PostForm("peerAddr")
	err := a.fileServer.Join(peerName, peerAddr)
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

func (a *ApiServer) NewBoard(ctx *gin.Context) {
	err := a.fileServer.Group.NewBorad(ctx.Param("key"))
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

func (a *ApiServer) MkDir(ctx *gin.Context) {
	key, _ := ctx.GetQuery("key")
	base, _ := ctx.GetQuery("base")
	if base == "root" || base == "" {
		base = "."
	}
	dirName, _ := ctx.GetQuery("dir")
	err := a.fileServer.Group.FrontSystem.MakeDir(key, base, dirName)
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

func (a *ApiServer) GetDir(ctx *gin.Context) {
	key, _ := ctx.GetQuery("key")
	base, _ := ctx.GetQuery("base")
	if base == "root" {
		base = "."
	}
	dirName, _ := ctx.GetQuery("dir")
	subs, err := a.fileServer.Group.FrontSystem.GetDirSub(key, base, dirName)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"msg":     err.Error(),
			"success": errors.Is(err, nil),
			"sub":     nil,
		})
	} else {
		ctx.JSON(http.StatusOK, gin.H{
			"msg":     "success",
			"success": errors.Is(err, nil),
			"sub":     subs,
		})
	}

}
