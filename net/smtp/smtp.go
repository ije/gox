package smtp

import (
	"errors"
	"fmt"
	netmail "net/mail"
	"net/smtp"
	"sync"
)

type SMTP struct {
	addr     string
	auth     smtp.Auth
	username string
}

func New(host string, port uint16, username string, password string) *SMTP {
	return &SMTP{
		addr:     fmt.Sprintf("%s:%d", host, port),
		auth:     smtp.PlainAuth("", username, password, host),
		username: username,
	}
}

func (s *SMTP) Auth() (err error) {
	c, err := smtp.Dial(s.addr)
	if err != nil {
		return
	}
	defer c.Close()
	return c.Auth(s.auth)
}

func (s *SMTP) SendMail(mail *Mail, from string, to string, oneToOne bool) (err error) {
	if mail == nil {
		err = errors.New("mail is nil")
		return
	}

	var sender *netmail.Address
	if from == "" {
		sender, err = netmail.ParseAddress(s.username)
	} else {
		sender, err = netmail.ParseAddress(from)
	}
	if err != nil {
		return
	}

	list, err := netmail.ParseAddressList(to)
	if err != nil {
		return
	}

	recipients := AddressList(list)

	if !oneToOne {
		err = smtp.SendMail(s.addr, s.auth, sender.Address, recipients.List(), mail.Encode(sender, recipients))
		if err != nil {
			err = &SendError{Message: err.Error(), From: sender, To: recipients}
		}
		return
	}

	var wg sync.WaitGroup
	var errs OTOSendError
	for _, recipient := range recipients {
		wg.Add(1)
		go func(to *netmail.Address) {
			err = smtp.SendMail(s.addr, s.auth, sender.Address, []string{to.Address}, mail.Encode(sender, AddressList{to}))
			if err != nil {
				errs.Errors = append(errs.Errors, &SendError{Message: err.Error(), From: sender, To: AddressList{to}})
			}
			wg.Done()
		}(recipient)
	}
	wg.Wait()
	if len(errs.Errors) > 0 {
		err = &errs
	}
	return
}
