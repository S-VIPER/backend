package email

import (
	"context"
	"fmt"
	"net/smtp"
	"strconv"

	"github.com/S-VIPER/backend/gin-api/internal/service"
)

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type SMTPSender struct {
	config SMTPConfig
}

func NewSMTPSender(config SMTPConfig) *SMTPSender {
	return &SMTPSender{
		config: config,
	}
}

var _ service.EmailSender = (*SMTPSender)(nil)

func (s *SMTPSender) SendRegistrationCode(
	ctx context.Context,
	to string,
	code string,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	const subject = "Verify your S-VIPER account"

	body := fmt.Sprintf(
		"Hello!\r\n"+
			"\r\n"+
			"Your S-VIPER verification code is:\r\n"+
			"\r\n"+
			"%s\r\n"+
			"\r\n"+
			"The code expires in 10 minutes.\r\n"+
			"\r\n"+
			"If you did not request this registration, you can ignore this email.\r\n"+
			"\r\n"+
			"S-VIPER\r\n",
		code,
	)

	message := fmt.Sprintf(
		"From: S-VIPER <%s>\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"Content-Type: text/plain; charset=UTF-8\r\n"+
			"\r\n"+
			"%s",
		s.config.From,
		to,
		subject,
		body,
	)

	address := s.config.Host + ":" + strconv.Itoa(s.config.Port)

	var auth smtp.Auth

	if s.config.Username != "" || s.config.Password != "" {
		auth = smtp.PlainAuth(
			"",
			s.config.Username,
			s.config.Password,
			s.config.Host,
		)
	}

	if err := smtp.SendMail(
		address,
		auth,
		s.config.From,
		[]string{to},
		[]byte(message),
	); err != nil {
		return fmt.Errorf("send registration code email: %w", err)
	}

	return nil
}
