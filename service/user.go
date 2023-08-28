package service

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"time"

	"github.com/ciiim/cloudborad/errmsg"
	"github.com/ciiim/cloudborad/models"
	"github.com/dlclark/regexp2"
)

// 密码检查函数
var (
	LowPWChecker = func(passwd string) bool {
		//密码大于等于8位，无其它限制
		return len(passwd) >= 8
	}
	MediumPWChecker = func(passwd string) bool {
		//密码大于等于8位，必须包含数字，小写字母和大写字母
		if !LowPWChecker(passwd) {
			return false
		}
		re := regexp2.MustCompile(`^(?=.*[0-9])(?=.*[a-z])(?=.*[A-Z]).+$`, 0)
		match, _ := re.MatchString(passwd)
		return match
	}
	HighPWChecker = func(passwd string) bool {
		//密码大于等于8位，必须包含数字，小写字母，大写字母和特殊字符
		if !LowPWChecker(passwd) {
			return false
		}
		re := regexp2.MustCompile(`^(?=.*[0-9])(?=.*[a-z])(?=.*[A-Z])(?=.*[!@#$%^&*()_+]).+$`, 0)
		match, _ := re.MatchString(passwd)
		return match
	}
)

var PasswordChecker = MediumPWChecker

func EncryptPasswd(passwd string) string {
	sum := sha1.Sum([]byte(passwd))
	return hex.EncodeToString(sum[:])
}

func VerifyCodeChecker(email, verifyCode string) bool {
	//redis get
	return true
}

func RegisterUser(email, nickName, passwd, phoneNumber, verifyCode string) (int64, error) {
	if !VerifyCodeChecker(email, verifyCode) {
		return -1, errmsg.ErrWrongVerifyCode
	}

	if !PasswordChecker(passwd) {
		return -1, errmsg.ErrWeakPasswd
	}
	encryptedPasswd := EncryptPasswd(passwd)
	newUser := &models.NewUser{
		Email:         email,
		NickName:      nickName,
		Passwd:        encryptedPasswd,
		PhoneNumber:   phoneNumber,
		RegisterTime:  time.Now(),
		AccountStatus: 0,
	}
	if uid, err := models.InsertUserBasic(newUser); errors.Is(err, errmsg.ErrInsertUserFailed) {
		return -1, errmsg.ErrUserExist
	} else {
		return uid, err
	}
}

func LoginByPasswd(email, passwd string) (Token, bool, error) {
	encryptedPasswd := EncryptPasswd(passwd)
	user, err := models.QueryUserBasic(email)
	if err != nil {
		if errors.Is(err, errmsg.ErrQueryUserFailed) {
			return Token{}, false, errmsg.ErrUserNotFound
		} else {
			return Token{}, false, err
		}
	}

	if user.Passwd != encryptedPasswd {
		return Token{}, false, errmsg.ErrWrongPasswd
	}
	token := GenerateToken(user.UID, user.NickName, user.PermissionGroup, ExpireTimeMonth)
	return token, true, nil
}
