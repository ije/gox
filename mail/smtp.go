package mail

import (
	"fmt"
	"net/smtp"
)

type Smtp struct {
	addr string
	auth smtp.Auth
}

func NewSmtp(host string, port int, username, password string) *Smtp {
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

func (server *Smtp) Send(from Contact, to Contacts, subject, text, html string) error {
	mail, err := NewMailBody(from, to, subject, text, html)
	if err != nil {
		return err
	}
	return smtp.SendMail(server.addr, server.auth, from.Addr, to.List(), mail.Body())
}
