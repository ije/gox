package smtp

import (
	"errors"
	"fmt"
	netmail "net/mail"
	"net/smtp"
	"strings"
	"sync"

	"github.com/ije/gox/valid"
)

type SMTP struct {
	DefaultFrom *netmail.Address
	addr        string
	auth        smtp.Auth
}

func New(host string, port uint16, username string, password string, defaultForm string) *SMTP {
	return &SMTP{Address(defaultForm), fmt.Sprintf("%s:%d", host, port), smtp.PlainAuth("", username, password, host)}
}

func (s *SMTP) Auth() (err error) {
	c, err := smtp.Dial(s.addr)
	if err != nil {
		return
	}
	defer c.Close()
	return c.Auth(s.auth)
}

func (s *SMTP) SendMail(mail *Mail, from interface{}, to interface{}, oneToOne bool) (err error) {
	if mail == nil {
		err = errors.New("mail is nil")
		return
	}

	var sender *netmail.Address
	var recipients AddressList

	if from != nil {
		switch a := from.(type) {
		case string:
			sender, err = netmail.ParseAddress(a)
			if err != nil {
				return
			}
		case netmail.Address:
			if valid.IsEmail(a.Address) {
				sender = &a
			}
		case *netmail.Address:
			if valid.IsEmail(a.Address) {
				sender = a
			}
		}
	}
	if sender == nil {
		sender = s.DefaultFrom
	}
	if sender == nil {
		err = ErrEmptySender
		return
	}

	if to != nil {
		switch a := to.(type) {
		case string:
			var list []*netmail.Address
			list, err = netmail.ParseAddressList(a)
			if err != nil {
				return
			}
			recipients = AddressList(list)
		case []netmail.Address:
			for _, s := range a {
				if valid.IsEmail(s.Address) {
					recipients = append(recipients, &s)
				}
			}
		case []*netmail.Address:
			for _, s := range a {
				if valid.IsEmail(s.Address) {
					recipients = append(recipients, s)
				}
			}
		case AddressList:
			for _, s := range a {
				if valid.IsEmail(s.Address) {
					recipients = append(recipients, s)
				}
			}
		case map[string]string:
			tmp := map[string]string{}
			for name, email := range a {
				if valid.IsEmail(email) {
					tmp[email] = strings.TrimSpace(name)
				}
			}
			for email, name := range tmp {
				recipients = append(recipients, &netmail.Address{Name: name, Address: email})
			}
		}
	}
	if recipients == nil || len(recipients) == 0 {
		err = ErrEmptyRecipients
		return
	}

	if len(mail.Subject) == 0 {
		err = ErrEmptySubject
		return
	}

	if len(mail.PlainText) == 0 && len(mail.Html) == 0 {
		err = ErrEmptyContent
		return
	}

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
