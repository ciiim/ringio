package router

import (
	"errors"
	"net/http"

	"github.com/ciiim/cloudborad/internal/fs"
	"github.com/ciiim/cloudborad/server"
	"github.com/gin-gonic/gin"
)

/*

返回格式规定
code: int
success: bool
msg: string
data: gin.H{} or nil
*/

const (
	apiVersion = "v1"
)

var (
	apiBasePath = "/api/" + apiVersion
)

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
	as := &ApiServer{
		r:          r,
		fileServer: fileServer,
	}

	APIGroup := r.Group(apiBasePath)
	{
		fsAPIGroup := APIGroup.Group("fs", jwtAuth())
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
			authAPIGroup.POST("/login", as.LoginByPasswd)

			//注册
			authAPIGroup.POST("/register", as.Register)

			//登出
			authAPIGroup.POST("/logout")

			//使用邮箱激活验证账号
			authAPIGroup.POST("/verify")
		}

		adminGroup := as.r.Group("/internal/admin", jwtAdminAuth())
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

func (a *ApiServer) Run(port string) {
	a.r.Run(":" + port)
}

func (a *ApiServer) GetCluster(ctx *gin.Context) {
	list := a.fileServer.Group.FrontSystem.Peer().PList()
	dpeerList := fs.PeerInfoListToDpeerInfoList(a.fileServer.Group.FrontSystem.Peer().PList())
	ctx.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"meg":     "success",
		"success": true,
		"data": gin.H{
			"peernum":  len(list),
			"peerlist": dpeerList,
		},
	})
}

func (a *ApiServer) QuitCluster(ctx *gin.Context) {
	a.fileServer.Quit()
	ctx.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"msg":     "success",
		"success": true,
		"data":    nil,
	})
}

func (a *ApiServer) JoinCluster(ctx *gin.Context) {
	peerName := ctx.PostForm("peerName")
	peerAddr := ctx.PostForm("peerAddr")
	err := a.fileServer.Join(peerName, peerAddr)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"msg":     err.Error(),
			"success": errors.Is(err, nil),
			"data":    nil,
		})
		return
	} else {
		ctx.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"msg":     "success",
			"success": errors.Is(err, nil),
			"data":    nil,
		})
	}
}

/*
Space API
*/

func (a *ApiServer) NewBoard(ctx *gin.Context) {
	err := a.fileServer.NewBoard(ctx.Param("key"))
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"msg":     err.Error(),
			"success": errors.Is(err, nil),
			"data":    nil,
		})
		return
	} else {
		ctx.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"msg":     "success",
			"success": errors.Is(err, nil),
			"data":    nil,
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
			"code":    http.StatusOK,
			"msg":     err.Error(),
			"success": errors.Is(err, nil),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
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
			"code":    http.StatusOK,
			"msg":     err.Error(),
			"success": errors.Is(err, nil),
			"data": gin.H{
				"sub": nil,
			},
		})
	} else {
		ctx.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"msg":     "success",
			"success": errors.Is(err, nil),
			"data": gin.H{
				"sub": subs,
			},
		})
	}

}
