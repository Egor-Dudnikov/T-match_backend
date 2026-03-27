package service

import (
	"T-match_backend/internal/models"
	"net/smtp"
	"os"
)

type EmailClient struct {
	cfg models.EmailConfig
}

func NewEmailClient(cfg models.EmailConfig) *EmailClient {
	return &EmailClient{cfg: cfg}
}

func (r *EmailClient) SendVerifyCode(to string, code string) error {
	addr := r.cfg.Addr
	a := smtp.PlainAuth(r.cfg.Identity, r.cfg.Username, os.Getenv("SMTP_PASSWORD"), r.cfg.Host)
	from := r.cfg.Username
	msg := []byte("From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: Code for verify\r\n" +
		"\r\n" +
		"Code:" + code + "\r\n")

	err := smtp.SendMail(addr, a, from, []string{to}, msg)
	if err != nil {
		return err
	}
	return nil
}
