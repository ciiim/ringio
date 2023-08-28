package router

import (
	"net/http"

	"github.com/ciiim/cloudborad/service"
	"github.com/gin-gonic/gin"
)

type loginByPasswdForm struct {
	Email  string `form:"email" json:"email" binding:"required"`
	Passwd string `form:"passwd" json:"passwd" binding:"required"`
}

type loginByCodeForm struct {
	Email      string `form:"email" json:"email" binding:"required"`
	VerifyCode string `form:"verify_code" json:"verify_code" binding:"required"`
}

type RegisterForm struct {
	Email       string `form:"email" json:"email" binding:"required"`
	NickName    string `form:"nickname" json:"nick_name" binding:"required"`
	Passwd      string `form:"passwd" json:"passwd" binding:"required"`
	PhoneNumber string `form:"phone_number" json:"phone_number"`
	VerifyCode  string `form:"verify_code" json:"verify_code" binding:"required"`
}

func (a *ApiServer) Register(c *gin.Context) {
	var form RegisterForm
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusOK,
			"success": false,
			"msg":     "参数错误",
			"data":    nil,
		})
		return
	}
	uid, err := service.RegisterUser(form.Email, form.NickName, form.Passwd, form.PhoneNumber, form.VerifyCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusOK,
			"success": false,
			"msg":     err.Error(),
			"data":    nil,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"success": true,
		"msg":     "注册成功",
		"data": gin.H{
			"uid": uid,
		},
	})
}

func (a *ApiServer) LoginByPasswd(c *gin.Context) {
	var form loginByPasswdForm
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusOK,
			"success": false,
			"msg":     "参数错误",
			"data":    nil,
		})
		return
	}
	token, ok, err := service.LoginByPasswd(form.Email, form.Passwd)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusOK,
			"success": false,
			"msg":     err.Error(),
			"data":    nil,
		})
		return
	} else {
		c.SetCookie("token", token.Token, int(token.Age.Seconds()), "/", "localhost", false, true)
		c.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"success": true,
			"msg":     "登录成功",
			"data": gin.H{
				"token": token,
			},
		})
		return
	}
}
