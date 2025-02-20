package utils

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

var smtpSubject = os.Getenv("SMTP_SENDER")

type SMTPConnector struct {
	client *gomail.Dialer
}

func CreateSMTPConnector() (*SMTPConnector, error) {
	smtpHost, smtpPort, smtpPassword := os.Getenv("SMTP_HOST"), os.Getenv("SMTP_PORT"), os.Getenv("SMTP_PASSWORD")

	if len(smtpSubject) == 0 {
		return nil, fmt.Errorf("no SMTP_SENDER, please check environmental variables")
	}

	if len(smtpHost) == 0 {
		return nil, fmt.Errorf("no SMTP_HOST, please check environmental variables")
	}

	if len(smtpPort) == 0 {
		return nil, fmt.Errorf("no SMTP_PORT, please check environmental variables")
	}

	if len(smtpPassword) == 0 {
		return nil, fmt.Errorf("no SMTP_PASSWORD, please check environmental variables")
	}

	formattedSmtpPort, err := strconv.Atoi(smtpPort)

	if err != nil {
		return nil, fmt.Errorf("error at Atoi: %s", err)
	}

	client := gomail.NewDialer(smtpHost, formattedSmtpPort, smtpSubject, smtpPassword)

	return &SMTPConnector{client: client}, nil
}

func (c *SMTPConnector) SendMail(ctx context.Context, email string, fileLink string) error {
	var err error

	message := gomail.NewMessage()

	message.SetHeader("From", fmt.Sprintf("Archiver App <%s>", smtpSubject))
	message.SetHeader("To", email)
	message.SetHeader("Subject", "Your archive link")
	message.SetBody("text/plain", fileLink)

	err = c.client.DialAndSend(message)

	if err != nil {
		return fmt.Errorf("error at DialAndSend: %s", err)
	}

	return nil
}
