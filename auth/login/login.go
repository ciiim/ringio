package login

import (
	"log"

	"github.com/ciiim/cloudborad/auth"
	dbo "github.com/ciiim/cloudborad/internal/database"
)

type Login struct {
	Data  LoginData
	token auth.Identify
}

type LoginData struct {
	uid      uint64 // unique user id
	username string
	passwd   string
}

func New(uid uint64, username string, passwd string) *Login {
	return &Login{
		Data: LoginData{
			uid:      uid,
			username: username,
			passwd:   passwd,
		},
		token: nil,
	}
}

func (l *Login) exist() bool {
	has, err := dbo.HasUser(l.Data.uid)
	if err != nil {
		log.Println(err)
		return false
	}
	return has
}

func (l *Login) Do() {
}

func (l *Login) createIdentity() auth.Identify {
	return auth.NewToken(l.Data.uid)
}
