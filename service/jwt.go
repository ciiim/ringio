package service

import (
	"fmt"
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

func VerifyToken(token string) (bool, error) {
	j, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("[VerifyToken] Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return false, err
	}
	if j.Valid {
		return true, nil
	}
	return false, nil
}

func VerifyAdmin(token string) (bool, error) {
	j, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return false, err
	}
	if j.Valid {
		if claims, ok := j.Claims.(jwt.MapClaims); ok {
			if claims["permissiongroup"].(int) == 1 {
				return true, nil
			}
		}
	}
	return false, nil
}
