package router

import "github.com/gin-gonic/gin"

func jwtAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
