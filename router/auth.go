package router

import (
	"net/http"

	"github.com/ciiim/cloudborad/errmsg"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

const (
	// 1000 - 1999 用户相关
	Login_Success            = 1000
	Login_FormErr            = 1001
	Login_WrongPasswdOrEmail = 1004
	Auth_Unauthorized        = 1005

	Reg_Success = 1100
	Reg_Failed  = 1101
	Reg_FormErr = 1102

	Reset_Success = 1300
	Reset_Failed  = 1301

	ResetEmailSend_Success = 1310
	ResetEmailSend_Failed  = 1311

	Verify_Success = 1200
	Verify_Failed  = 1201
	Verify_FormErr = 1202
)

func (a *ApiServer) Verify(c *gin.Context) {
	token := GetToken(c)
	j, err := a.service.ParseToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    Auth_Unauthorized,
			"success": false,
			"msg":     errmsg.ErrMsgCN(err),
			"data":    nil,
		})
		return
	}
	ok, err := a.service.VerifyToken(j)
	if !ok || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    Auth_Unauthorized,
			"success": false,
			"msg":     errmsg.ErrMsgCN(err),
			"data":    nil,
		})
		return
	}
}

type verifyForm struct {
	Email string `form:"email" json:"email" binding:"required"`
	Type  string `form:"type" json:"type" binding:"required"`
}

func (a *ApiServer) SendVerifyCodeEmail(c *gin.Context) {
	var form verifyForm
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    Verify_FormErr,
			"success": false,
			"msg":     "参数错误",
			"data":    nil,
		})
		return
	}
	expireTime, err := a.service.SendVerifyToEmail(form.Email, form.Type)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    Verify_Failed,
			"success": false,
			"msg":     errmsg.ErrMsgCN(err) + "real error:" + err.Error(),
			"data":    nil,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    Verify_Success,
		"success": true,
		"msg":     "验证码发送成功",
		"data": gin.H{
			"expire_time": expireTime.Seconds(),
		},
	})

}

type RegisterForm struct {
	Email       string `form:"email" json:"email" binding:"required"`
	NickName    string `form:"nickname" json:"nickname" binding:"required"`
	Passwd      string `form:"passwd" json:"passwd" binding:"required"`
	PhoneNumber string `form:"phone_number" json:"phone_number"`
	VerifyCode  string `form:"verify_code" json:"verify_code" binding:"required"`
	VerifyType  string `form:"verify_type" json:"verify_type" binding:"required"`
}

func (a *ApiServer) Register(c *gin.Context) {
	var form RegisterForm
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    Reg_FormErr,
			"success": false,
			"msg":     "参数错误",
			"data":    nil,
		})
		return
	}
	uid, err := a.service.RegisterUser(form.Email, form.NickName, form.Passwd, form.PhoneNumber, form.VerifyCode, form.VerifyType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    Reg_Failed,
			"success": false,
			"msg":     errmsg.ErrMsgCN(err),
			"data":    nil,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    Reg_Success,
		"success": true,
		"msg":     "注册成功",
		"data": gin.H{
			"uid": uid,
		},
	})
}

type loginTypeForm struct {
	Type string `form:"type" json:"type" binding:"required"`
}

func (a *ApiServer) Login(c *gin.Context) {
	var form loginTypeForm
	if err := c.ShouldBindBodyWith(&form, binding.JSON); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    Login_FormErr,
			"success": false,
			"msg":     "参数错误" + err.Error(),
			"data":    nil,
		})
		return
	}
	if form.Type == "passwd" {
		a.loginByPasswd(c)
	} else if form.Type == "code" {
		a.loginByCode(c)
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    Login_FormErr,
			"success": false,
			"msg":     "参数错误",
			"data":    nil,
		})
		return
	}
}

type loginByPasswdForm struct {
	Email  string `form:"email" json:"email" binding:"required"`
	Passwd string `form:"passwd" json:"passwd" binding:"required"`
}

func (a *ApiServer) loginByPasswd(c *gin.Context) {
	var form loginByPasswdForm
	if err := c.ShouldBindBodyWith(&form, binding.JSON); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    Login_FormErr,
			"success": false,
			"msg":     "参数错误",
			"data":    nil,
		})
		return
	}
	token, ok, err := a.service.LoginByPasswd(form.Email, form.Passwd)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    Login_WrongPasswdOrEmail,
			"success": false,
			"msg":     errmsg.ErrMsgCN(err),
			"data":    nil,
		})
		return
	} else {
		c.Header("Access-Control-Allow-Origin", c.Request.Header.Get("Origin"))
		c.Header("Authorization", token.Token)
		c.JSON(http.StatusOK, gin.H{
			"code":    Login_Success,
			"success": true,
			"msg":     "登录成功",
			"data": gin.H{
				"token": token.Token,
			},
		})
		return
	}
}

type loginByCodeForm struct {
	Email      string `form:"email" json:"email" binding:"required"`
	VerifyCode string `form:"verify_code" json:"verify_code" binding:"required"`
	VerifyType string `form:"verify_type" json:"verify_type" binding:"required"`
}

func (a *ApiServer) loginByCode(c *gin.Context) {
	var form loginByCodeForm
	if err := c.ShouldBindBodyWith(&form, binding.JSON); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    Login_FormErr,
			"success": false,
			"msg":     "参数错误code" + err.Error(),
			"data":    nil,
		})
		return
	}
	token, ok, err := a.service.LoginByCode(form.Email, form.VerifyCode, form.VerifyType)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    Login_WrongPasswdOrEmail,
			"success": false,
			"msg":     errmsg.ErrMsgCN(err),
			"data":    nil,
		})
		return
	} else {
		c.Header("Access-Control-Allow-Origin", c.Request.Header.Get("Origin"))
		c.Header("Authorization", token.Token)
		c.JSON(http.StatusOK, gin.H{
			"code":    Login_Success,
			"success": true,
			"msg":     "登录成功",
			"data": gin.H{
				"token": token.Token,
			},
		})
		return
	}
}

type sendVerifyEmailForm struct {
	Email string `form:"email" json:"email" binding:"required"`
}

func (a *ApiServer) SendResetEmail(c *gin.Context) {
	host := c.Request.Header.Get("Origin") + "/#"
	var form sendVerifyEmailForm
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    Verify_FormErr,
			"success": false,
			"msg":     "参数错误",
			"data":    nil,
		})
		return
	}
	err := a.service.SendResetEmail(host, form.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    ResetEmailSend_Failed,
			"success": false,
			"msg":     errmsg.ErrMsgCN(err),
			"data":    nil,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    ResetEmailSend_Success,
		"success": true,
		"msg":     "重置邮件发送成功",
		"data":    nil,
	})
}

type resetPasswdForm struct {
	Email      string `form:"email" json:"email" binding:"required"`
	NewPasswd  string `form:"new_passwd" json:"new_passwd" binding:"required"`
	ResetToken string `form:"token" json:"token" binding:"required"`
}

func (a *ApiServer) ResetPasswd(c *gin.Context) {
	var form resetPasswdForm
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    Reg_FormErr,
			"success": false,
			"msg":     "参数错误",
			"data":    nil,
		})
		return
	}
	err := a.service.ResetPasswd(form.Email, form.NewPasswd, form.ResetToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    Reset_Failed,
			"success": false,
			"msg":     errmsg.ErrMsgCN(err),
			"data":    nil,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    Reset_Success,
		"success": true,
		"msg":     "修改成功",
		"data":    nil,
	})
}

type checkResetTokenForm struct {
	Token string `form:"token" json:"token" binding:"required"`
}

func (a *ApiServer) CheckResetToken(c *gin.Context) {
	var form checkResetTokenForm
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    Verify_FormErr,
			"success": false,
			"msg":     "参数错误",
			"data":    nil,
		})
	}
	j, err := a.service.ParseToken(form.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    Verify_Failed,
			"success": false,
			"msg":     errmsg.ErrMsgCN(err),
			"data":    nil,
		})
	}
	ok, err := a.service.VerifyResetToken(j)
	if !ok || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    Verify_Failed,
			"success": false,
			"msg":     errmsg.ErrMsgCN(err),
			"data":    nil,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    Verify_Success,
		"success": true,
		"msg":     "验证成功",
		"data":    nil,
	})
}
