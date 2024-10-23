package email

import (
	"fmt"
	"net/smtp"
	"strings"

	"github.com/Alsond5/StockMarketAPIWebScraper/internal/logger"
)

type Configurations struct {
	Host     string
	Port     int
	Username string
	Password string
}

// Email struct
type Email struct {
	To      string
	Subject string
	Headers []string
	Body    string
}

func NewConfigurations(host string, port int, username, password string) *Configurations {
	return &Configurations{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
	}
}

// NewEmail creates a new Email instance
func NewEmail(to, subject, body string, headers []string) *Email {
	return &Email{
		To:      to,
		Subject: subject,
		Headers: headers,
		Body:    body,
	}
}

// Send sends an email :params config: Configurations
func (e *Email) Send(config *Configurations) error {
	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)

	message := []byte("From: " + config.Username + "\r\n" +
		"To: " + e.To + "\r\n" +
		"Subject: " + e.Subject + "\r\n" +
		strings.Join(e.Headers, "\r\n") +
		"\r\n" +
		e.Body)

	err := smtp.SendMail(fmt.Sprintf("%s:%d", config.Host, config.Port), auth, config.Username, []string{e.To}, message)
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	logger.Success(fmt.Sprintf("Email has been sent to %s successfully.", e.To))

	return nil
}
