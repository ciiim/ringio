package router

import (
	"log"

	"github.com/ciiim/cloudborad/internal/fs"
	"github.com/ciiim/cloudborad/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

/*

返回格式规定
code: int
success: bool
msg: string
data: gin.H{} or nil
*/

var JSON_RETURN = func(code int, success bool, msg string, data gin.H) gin.H {
	return gin.H{
		"code":    code,
		"success": success,
		"msg":     msg,
		"data":    data,
	}
}

const (
	apiVersion = "v1"
)

var (
	apiBasePath = "/api/" + apiVersion
)

type ApiServer struct {
	r       *gin.Engine
	service *service.Service
}

func InitApiServer(service *service.Service) *ApiServer {
	if fs.IsDebug() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"PUT", "PATCH", "POST", "GET", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))
	as := &ApiServer{
		r:       r,
		service: service,
	}

	APIGroup := r.Group(apiBasePath)
	{
		fsAPIGroup := APIGroup.Group("fs", as.jwtAuth())
		{
			fsAPIGroup.GET("/board/list", as.GetAllBoardBasic)

			fsAPIGroup.GET("/board/sub", as.GetBoardSub)

			fsAPIGroup.POST("/board/:space", as.MakeDir)
			fsAPIGroup.PUT("/board/:space", as.RenameDir)
			fsAPIGroup.DELETE("/board/:space", as.DeleteDir)

			fsAPIGroup.POST("/board", as.NewBoard)
			fsAPIGroup.PUT("/board", as.UpdateBoard)
			fsAPIGroup.DELETE("/board", as.DeleteBoard)

		}

		authAPIGroup := APIGroup.Group("auth")
		{
			//登录
			authAPIGroup.POST("/login", as.Login)

			//注册
			authAPIGroup.POST("/register", as.Register)

			//登出
			authAPIGroup.POST("/logout")

			//修改密码

			authAPIGroup.POST("/reset-send", as.SendResetEmail)

			authAPIGroup.POST("/reset", as.ResetPasswd)

			authAPIGroup.POST("/check-reset-token", as.CheckResetToken)

			//发送验证码
			authAPIGroup.POST("/email-verify-code", as.SendVerifyCodeEmail)
		}

		adminGroup := as.r.Group("/admin", as.jwtAdminAuth())
		{
			as.r.LoadHTMLGlob("server/admin/*")

			//简易节点操作
			adminGroup.GET("/peer", as.AdminPage)

			adminGroup.GET("/cluster", as.GetCluster)
			adminGroup.POST("/cluster", as.JoinCluster)
			adminGroup.DELETE("/cluster", as.QuitCluster)
		}
	}

	return as
}

func (a *ApiServer) Run(port string) {
	if err := a.r.Run(":" + port); err != nil {
		log.Printf("[ApiServer] Run failed: %v", err)
	}
}
