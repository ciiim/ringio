package service

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/rand"
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

func encryptPasswd(passwd string) string {
	sum := sha1.Sum([]byte(passwd))
	return hex.EncodeToString(sum[:])
}

func (s *Service) VerifyCodeChecker(email, verifyCode, verifyType string) bool {
	if code, err := models.GetVerifyCode(email + verifyType); err != nil {
		return false
	} else if code == verifyCode {
		return true
	}
	return false
}

func (s *Service) EmailExist(email string) (bool, error) {
	return models.EmailExist(email)
}

func (s *Service) RegisterUser(email, nickName, passwd, phoneNumber, verifyCode, verifyType string) (int64, error) {
	if exist, _ := s.EmailExist(email); exist {
		return -1, errmsg.ErrUserExist
	}

	if !s.VerifyCodeChecker(email, verifyCode, verifyType) {
		return -1, errmsg.ErrWrongVerifyCode
	}

	if !PasswordChecker(passwd) {
		return -1, errmsg.ErrWeakPasswd
	}
	encryptedPasswd := encryptPasswd(passwd)
	newUser := &models.NewUser{
		Email:         email,
		NickName:      nickName,
		Passwd:        encryptedPasswd,
		PhoneNumber:   phoneNumber,
		RegisterTime:  time.Now(),
		AccountStatus: 0,
	}
	if uid, err := models.InsertUserBasic(newUser); errors.Is(err, errmsg.ErrInsertUserFailed) {
		models.DeleteVerifyCode(email + verifyType)
		return -1, errmsg.ErrUserExist
	} else {
		models.DeleteVerifyCode(email + verifyType)
		return uid, err
	}
}

func (s *Service) LoginByPasswd(email, passwd string) (Token, bool, error) {
	encryptedPasswd := encryptPasswd(passwd)
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
	token := generateToken(user.UID, user.NickName, user.PermissionGroup, ExpireTimeMonth)
	return token, true, nil
}

func (s *Service) LoginByCode(email, verifyCode, verifyType string) (Token, bool, error) {
	if !s.VerifyCodeChecker(email, verifyCode, verifyType) {
		return Token{}, false, errmsg.ErrWrongVerifyCode
	}

	user, err := models.QueryUserBasic(email)
	if err != nil {
		if errors.Is(err, errmsg.ErrQueryUserFailed) {
			return Token{}, false, errmsg.ErrUserNotFound
		} else {
			return Token{}, false, err
		}
	}

	token := generateToken(user.UID, user.NickName, user.PermissionGroup, ExpireTimeMonth)
	//delete verify code
	models.DeleteVerifyCode(email + verifyType)

	return token, true, nil
}

func (s *Service) SendVerifyToEmail(email, codeType string) (time.Duration, error) {
	//check email exist in redis
	if code, err := models.GetVerifyCode(email + codeType); err == nil && code != "" {
		return 0, errmsg.ErrVerifyCodeExist
	} else if err != nil {
		return 0, err
	}
	//generate verify code
	code := generateVerifyCode(s.email.VerifyCodeLen)
	log.Printf("verify code: %s", code)
	//send verify code
	err := s.sendVerifyCode(email, code)
	if err != nil {
		return 0, errmsg.ErrSendVerifyCodeFailed
	}
	//redis set
	models.SetVerifyCode(email, codeType, code, s.email.VerifyCodeExpireTime)

	return s.email.VerifyCodeExpireTime, nil
}

func generateVerifyCode(length int) string {
	code := ""
	for i := 0; i < length; i++ {
		code += fmt.Sprintf("%d", rand.Intn(10))
	}
	return code
}

func (s *Service) SendResetEmail(frontUrl, email string) error {
	if exist, _ := s.EmailExist(email); !exist {
		return errmsg.ErrUserNotFound
	}
	token := generateResetToken(email)
	err := s.sendResetTokenEmail(frontUrl, email, token.Token, token.Age)
	if err != nil {
		return errmsg.ErrSendVerifyCodeFailed
	}
	return nil
}

func (s *Service) ResetPasswd(email, newPasswd, resetToken string) error {
	rj, err := s.ParseToken(resetToken)
	if err != nil {
		log.Printf("[ResetPasswd] ParseToken failed: %v", err)
		return errmsg.ErrResetTokenInvalid
	}
	if ok, err := s.VerifyResetToken(rj); !ok {
		log.Printf("[ResetPasswd] VerifyResetToken failed: %v", err)
		return errmsg.ErrResetTokenInvalid
	}
	if !PasswordChecker(newPasswd) {
		return errmsg.ErrWeakPasswd
	}
	encryptedPasswd := encryptPasswd(newPasswd)
	if err := models.UpdateUserPasswd(email, encryptedPasswd); err != nil {
		return err
	}
	return nil
}
