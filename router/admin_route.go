package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *ApiServer) AdminPage(ctx *gin.Context) {
	name, ip := a.service.ServerInfo()
	dpeerList := a.service.GetClusterList()
	ctx.HTML(http.StatusOK, "peer.html", gin.H{
		"peerName": name,
		"peerAddr": ip,
		"peerList": dpeerList,
	})
}
