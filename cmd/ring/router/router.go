package router

import (
	"github.com/ciiim/cloudborad/cmd/ring/service"
	"github.com/gin-gonic/gin"
)

func Router() *gin.Engine {
	g := gin.Default()
	initRouter(g)
	return g
}

func initRouter(g *gin.Engine) {
	g.GET("/filecontent/:space/:filehash", service.GetFileContent)

	g.POST("/space/:space", service.GetSpaceWithDir)

	// 上传文件
	g.POST("/file/upload/:space", service.UploadFile)

	// 创建space
	g.POST("/space", service.CreateSpace)

	// // 删除space
	// g.DELETE("/space/:space", service.DeleteSpace)

	// // 创建文件夹
	// g.POST("/dir/:space", service.CreateDir)

	// // 删除文件夹
	// g.DELETE("/dir/:space", service.DeleteDir)

	// // 删除文件
	// g.DELETE("/file/:space", service.DeleteFile)

	// // 重命名文件
	// g.PUT("/file/:space", service.RenameFile)

	// // 重命名文件夹
	// g.PUT("/dir/:space", service.RenameDir)

}
