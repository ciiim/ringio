package service

import (
	"log"
	"time"

	"github.com/ciiim/cloudborad/errmsg"
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

func generateToken(uid int64, nickName string, permissionGroup int, expDeltaTime time.Duration) Token {
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

func generateResetToken(email string) Token {
	j := jwt.New(jwt.SigningMethodHS256)
	nowTime := time.Now()
	j.Claims = jwt.MapClaims{
		"iat":   nowTime.Unix(),
		"exp":   nowTime.Add(10 * time.Minute).Unix(),
		"email": email,
	}
	token, _ := j.SignedString([]byte(secret))
	return Token{
		Token: token,
		Age:   10 * time.Minute,
	}
}

func (s *Service) ParseToken(token string) (*jwt.Token, error) {
	j, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	return j, err
}

func (s *Service) VerifyResetToken(j *jwt.Token) (bool, error) {
	ok, err := s.VerifyToken(j)
	if !ok {
		log.Printf("[VerifyResetToken] VerifyToken failed: %v", err)
		return ok, errmsg.ErrResetTokenInvalid
	}
	return ok, nil
}

func (s *Service) VerifyToken(j *jwt.Token) (bool, error) {
	if j == nil {
		return false, errmsg.ErrTokenInvalid
	}
	if !j.Valid {
		return false, errmsg.ErrTokenInvalid
	}
	return true, nil
}

func (s *Service) VerifyAdmin(j *jwt.Token) (bool, error) {
	if j == nil {
		return false, errmsg.ErrTokenInvalid
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

func (s *Service) getTokenUID(j *jwt.Token) int64 {
	if claims, ok := j.Claims.(jwt.MapClaims); ok {
		return claims["uid"].(int64)
	}
	return -1
}
