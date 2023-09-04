package router

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetToken(c *gin.Context) string {
	token := c.GetHeader("Authorization")
	//分离Barer和token
	if len(token) > 7 && strings.ToUpper(token[0:7]) == "BEARER " {
		token = token[7:]
	}
	return token
}

func (a *ApiServer) jwtAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := GetToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    Auth_Unauthorized,
				"success": false,
				"msg":     "未登录",
				"data":    nil,
			})
			return
		}
		j, err := a.service.ParseToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    Auth_Unauthorized,
				"success": false,
				"msg":     "未登录" + err.Error(),
				"data":    nil,
			})
			return
		}
		if ok, _ := a.service.VerifyToken(j); !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    Auth_Unauthorized,
				"success": false,
				"msg":     "未登录" + err.Error(),
				"data":    nil,
			})
			return
		}
		c.Next()
	}
}

func (a *ApiServer) jwtAdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := GetToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    Auth_Unauthorized,
				"success": false,
				"msg":     "未登录",
				"data":    nil,
			})
			return
		}
		j, err := a.service.ParseToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    Auth_Unauthorized,
				"success": false,
				"msg":     "未登录",
				"data":    nil,
			})
			return
		}
		if ok, _ := a.service.VerifyAdmin(j); !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    Auth_Unauthorized,
				"success": false,
				"msg":     "未登录",
				"data":    nil,
			})
			return
		}
		c.Next()
	}
}
