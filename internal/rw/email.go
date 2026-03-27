package rw

import (
	"net/smtp"
	"os"
)

func SendVerifyCode(to string, code string, cfg VeryfyConfig) error {
	addr := cfg.Addr
	a := smtp.PlainAuth(cfg.Identity, cfg.Username, os.Getenv("SMTP_PASSWORD"), cfg.Host)
	from := cfg.Username
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
