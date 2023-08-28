package service

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Token struct {
	Token string
	Age   time.Duration
}

const (
	ExpireTimeMonth = 30 * 24 * time.Hour
)

const secret = "1#aM%A*sEcrEt"

func GenerateToken(uid int64, nickName string, permissionGroup int, expDeltaTime time.Duration) Token {
	j := jwt.New(jwt.SigningMethodHS256)
	nowTime := time.Now()
	j.Claims = jwt.MapClaims{
		"iat":             nowTime.Unix(),
		"exp":             nowTime.Add(expDeltaTime).Unix(),
		"uid":             uid,
		"nickname":        nickName,
		"permissiongroup": permissionGroup,
	}
	token, _ := j.SignedString([]byte(secret))
	return Token{
		Token: token,
		Age:   expDeltaTime,
	}
}
