package smtp

import (
	"fmt"
	"net/smtp"
)

type Smtp struct {
	addr string
	auth smtp.Auth
}

func New(host string, port int, username, password string) *Smtp {
	return &Smtp{fmt.Sprintf("%s:%d", host, port), smtp.PlainAuth("", username, password, host)}
}

func (s *Smtp) Auth() (err error) {
	c, err := smtp.Dial(s.addr)
	if err != nil {
		return
	}
	defer c.Close()
	return c.Auth(s.auth)
}

func (s *Smtp) SendMail(from, to interface{}, subject, text, html string, attachments []Attachment) error {
	mail, err := NewMail(from, to, subject, text, html, attachments)
	if err != nil {
		return err
	}
	return mail.Send(s)
}
