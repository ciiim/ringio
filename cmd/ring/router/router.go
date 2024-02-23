package router

import (
	"github.com/ciiim/cloudborad/cmd/ring/service"
	"github.com/gin-gonic/gin"
)

func Router() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	g := gin.New()
	g.Use(gin.Logger(), gin.Recovery())
	initRouter(g)
	return g
}

func initRouter(g *gin.Engine) {

	fileGroup := g.Group("/file")
	{
		fileGroup.GET("/:space/*path", service.GetFileContent)

		// 上传文件
		fileGroup.POST("/:space/upload", service.UploadFile)

		//删除文件
		fileGroup.DELETE("/:space/*path", service.DeleteFile)
	}

	spaceGroup := g.Group("/space")
	{
		// 创建space
		spaceGroup.POST("/", service.CreateSpace)

		spaceGroup.POST("/:space", service.GetSpaceWithDir)

	}

	// // 删除space
	// g.DELETE("/space/:space", service.DeleteSpace)

	// // 创建文件夹
	// g.POST("/dir/:space", service.CreateDir)

	// // 删除文件夹
	// g.DELETE("/dir/:space", service.DeleteDir)

	// // 重命名文件
	// g.PUT("/file/:space", service.RenameFile)

	// // 重命名文件夹
	// g.PUT("/dir/:space", service.RenameDir)

}
