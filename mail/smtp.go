package mail

/*
|--------------------------------------------------------------------------
| Smtp
|--------------------------------------------------------------------------
|
| Implements Smtp as part of the mail package in the Gofreight framework.
| Key symbols: SMTPMailer, NewSMTP, Send.
| 
| Mail covers Message and Mailer interfaces, LogMailer for tests,
| SMTP/SendGrid transports, mailable GFT rendering, and queued delivery.
| 
| Mailables render app/views/mail templates through the view engine;
| preview with gofreight mail:preview.
| 
| Configure MAIL_DRIVER in .env; authentication flows accept mail
| callbacks for reset and verification emails.
| 
| Symbols defined here include: SMTPMailer (exported type); NewSMTP
| (NewSMTP creates an SMTP mailer.).
| 
*/

import (
	"fmt"
	"net/smtp"
	"strings"
)

// SMTPMailer sends email via SMTP.
type SMTPMailer struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

// NewSMTP creates an SMTP mailer.
func NewSMTP(host, port, username, password, from string) *SMTPMailer {
	if port == "" {
		port = "587"
	}
	return &SMTPMailer{Host: host, Port: port, Username: username, Password: password, From: from}
}

func (m *SMTPMailer) Send(msg Message) error {
	from := msg.From
	if from == "" {
		from = m.From
	}
	body := msg.Body
	contentType := "text/plain; charset=UTF-8"
	if msg.HTML != "" {
		body = msg.HTML
		contentType = "text/html; charset=UTF-8"
	}

	headers := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: %s\r\n\r\n",
		from, strings.Join(msg.To, ", "), msg.Subject, contentType)
	payload := headers + body

	addr := fmt.Sprintf("%s:%s", m.Host, m.Port)
	auth := smtp.PlainAuth("", m.Username, m.Password, m.Host)
	return smtp.SendMail(addr, auth, from, msg.To, []byte(payload))
}
