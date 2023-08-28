package router

import (
	"errors"
	"net/http"

	"github.com/ciiim/cloudborad/service"
	"github.com/gin-gonic/gin"
)

func jwtAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Request.Cookie("token")
		if errors.Is(err, http.ErrNoCookie) || token == nil {
			c.Redirect(302, "/login")
			return
		}
		if ok, _ := service.VerifyToken(token.Value); !ok {
			c.Redirect(302, "/login")
			return
		}
		c.Next()
	}
}

func jwtAdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Request.Cookie("token")
		if errors.Is(err, http.ErrNoCookie) || token == nil {
			c.Redirect(302, "/login")
			return
		}
		if ok, _ := service.VerifyToken(token.Value); !ok {
			c.Redirect(302, "/login")
			return
		}
		if ok, _ := service.VerifyAdmin(token.Value); !ok {
			c.Redirect(302, "/login")
			return
		}
		c.Next()
	}
}
