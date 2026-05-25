// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

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

	from := r.cfg.Username
	msg := []byte("From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: Code for verify\r\n" +
		"\r\n" +
		"Code:" + code + "\r\n")
	var err error
	password := os.Getenv("SMTP_PASSWORD")
	if len(password) == 0 {
		err = smtp.SendMail(addr, nil, from, []string{to}, msg)
	} else {
		a := smtp.PlainAuth(r.cfg.Identity, r.cfg.Username, password, r.cfg.Host)
		err = smtp.SendMail(addr, a, from, []string{to}, msg)
	}

	if err != nil {
		return err
	}
	return nil
}
