package errmsg

import "fmt"

var (
	ErrUserNotFound         = fmt.Errorf("user not found")
	ErrWrongPasswd          = fmt.Errorf("wrong passwd")
	ErrWeakPasswd           = fmt.Errorf("weak passwd")
	ErrWrongVerifyCode      = fmt.Errorf("wrong verify code")
	ErrUserExist            = fmt.Errorf("user exist")
	ErrVerifyCodeExist      = fmt.Errorf("verify code exist")
	ErrSendVerifyCodeFailed = fmt.Errorf("send verify code failed")
	ErrTokenInvalid         = fmt.Errorf("token invalid")
	ErrResetTokenInvalid    = fmt.Errorf("reset token invalid")
)

var (
	ErrUserNotFoundCN         = "用户不存在"
	ErrWrongPasswdCN          = "密码错误"
	ErrWeakPasswdCN           = "密码强度不够"
	ErrWrongVerifyCodeCN      = "验证码过期或错误"
	ErrUserExistCN            = "该邮箱已存在"
	ErrVerifyCodeExistCN      = "验证码仍在有效期内"
	ErrSendVerifyCodeFailedCN = "发送验证码失败"
	ErrTokenInvalidCN         = "token无效"
	ErrResetTokenInvalidCN    = "重置密码token无效"
)

func ErrMsgCN(err error) string {
	switch err {
	case ErrUserNotFound:
		return ErrUserNotFoundCN
	case ErrWrongPasswd:
		return ErrWrongPasswdCN
	case ErrWeakPasswd:
		return ErrWeakPasswdCN
	case ErrWrongVerifyCode:
		return ErrWrongVerifyCodeCN
	case ErrUserExist:
		return ErrUserExistCN
	case ErrVerifyCodeExist:
		return ErrVerifyCodeExistCN
	case ErrSendVerifyCodeFailed:
		return ErrSendVerifyCodeFailedCN
	case ErrTokenInvalid:
		return ErrTokenInvalidCN
	case ErrResetTokenInvalid:
		return ErrResetTokenInvalidCN
	default:
		return "未知错误"
	}
}
