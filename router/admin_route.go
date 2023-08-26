package router

import (
	"net/http"

	"github.com/ciiim/cloudborad/internal/fs"
	"github.com/gin-gonic/gin"
)

func adminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

func (a *ApiServer) Shutdown(ctx *gin.Context) {
	err := a.fileServer.Close()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg":     err.Error(),
			"success": false,
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"msg":     "server shutdown",
		"success": true,
	})
}

func (a *ApiServer) AdminPage(ctx *gin.Context) {
	name, ip := a.fileServer.ServerInfo()
	dpeerList := fs.PeerInfoListToDpeerInfoList(a.fileServer.Group.FrontSystem.Peer().PList())
	ctx.HTML(http.StatusOK, "peer.html", gin.H{
		"peerName": name,
		"peerAddr": ip,
		"peerList": dpeerList,
	})
}
