// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package service

import (
	"T-match_backend/internal/apierrors"
	"T-match_backend/internal/constants"
	"T-match_backend/internal/models"
	"fmt"
	"net/smtp"
	"os"
	"time"
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

	msg := []byte(fmt.Sprintf(
		"From: noreply@tmatch.space\r\n"+
			"To: %s\r\n"+
			"Subject: Подтверждение регистрации на T-match\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n\r\n"+
			"<div style='font-family: Arial; max-width: 500px;'>"+
			"<h2 style='color:#1a73e8;'>T-match</h2>"+
			"<p>Ваш код подтверждения:</p>"+
			"<div style='background:#f0f4ff;border:2px dashed #FFD700;border-radius:8px;padding:20px;text-align:center;font-size:32px;letter-spacing:8px;color:#1a73e8;font-weight:bold;'>%s</div>"+
			"<p style='color:#000;font-size:14px;'>Действителен %d минут</p>"+
			"</div>",
		to, code, constants.VerifyCodeTimeLife/time.Minute,
	))
	var err error
	password := os.Getenv("EMAIL_PASSWORD")
	if len(password) == 0 {
		err = smtp.SendMail(addr, nil, from, []string{to}, msg)
	} else {
		a := smtp.PlainAuth(r.cfg.Identity, r.cfg.Username, password, r.cfg.Host)
		err = smtp.SendMail(addr, a, from, []string{to}, msg)
	}

	if err != nil {
		return apierrors.Wrap(apierrors.ErrEmailSendFailed, err)
	}
	return nil
}
