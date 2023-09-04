package service

import (
	"time"

	"github.com/ciiim/cloudborad/server"
)

type EmailConfig struct {
	Smtp                 *SmtpConfig
	VerifyCodeLen        int
	VerifyCodeExpireTime time.Duration
}
type SmtpConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Service struct {
	fileServer *server.Server
	email      EmailConfig
}

func NewService(server *server.Server) *Service {
	return &Service{
		fileServer: server,
	}
}

func (s *Service) SetEmailConfig(config EmailConfig) {
	s.email = config
}
