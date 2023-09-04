package service

import (
	"context"
	"log"
	"net/smtp"
	"time"
)

func (s *Service) sendVerifyCode(email string, code string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	smtpConfig := s.email.Smtp
	auth := smtp.PlainAuth("", smtpConfig.Email, smtpConfig.Password, smtpConfig.Host)
	to := []string{email}
	msg := []byte("To: " + email + "\r\n" +
		"Subject: Cloud Board 验证码\r\n" +
		"\r\n" +
		"验证码：" + code + "\r\n")
	doneChan := make(chan struct{})
	go func() {
		err := smtp.SendMail(smtpConfig.Host+":"+smtpConfig.Port, auth, smtpConfig.Email, to, msg)
		doneChan <- struct{}{}
		if err != nil {
			log.Printf("[sendVerifyCode] SendMail failed: %v", err)
		}
	}()
	select {
	case <-ctx.Done():
		log.Printf("[sendVerifyCode] SendMail timeout: %v", ctx.Err())
		return ctx.Err()
	case <-doneChan:
	}
	return nil
}

func (s *Service) sendResetTokenEmail(frontUrl, email, token string, expireTime time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	smtpConfig := s.email.Smtp
	auth := smtp.PlainAuth("", smtpConfig.Email, smtpConfig.Password, smtpConfig.Host)
	to := []string{email}
	msg := []byte("To: " + email + "\r\n" +
		"Subject: Cloud Board 重置密码\r\n" +
		"\r\n" +
		"重置密码链接：" + frontUrl + "/reset_password/" + email + "/" + token + "\r\n" +
		"链接有效时间：" + expireTime.String() + "\r\n")
	doneChan := make(chan struct{})
	go func() {
		err := smtp.SendMail(smtpConfig.Host+":"+smtpConfig.Port, auth, smtpConfig.Email, to, msg)
		doneChan <- struct{}{}
		if err != nil {
			log.Printf("[sendResetTokenEmail] SendMail failed: %v", err)
		}
	}()
	select {
	case <-ctx.Done():
		log.Printf("[sendResetTokenEmail] SendMail timeout: %v", ctx.Err())
		return ctx.Err()
	case <-doneChan:
	}
	return nil
}
