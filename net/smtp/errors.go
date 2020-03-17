package smtp

import (
	"net/mail"
	"strings"
)

// SendError records send a mail error
type SendError struct {
	Message string
	From    *mail.Address
	To      AddressList
}

func (err *SendError) Error() string {
	return err.Message
}

// SendErrors records send mails error
type SendErrors struct {
	Errors []*SendError
}

func (err *SendErrors) Error() string {
	var es []string
	for _, e := range err.Errors {
		es = append(es, e.Error())
	}
	return strings.Join(es, ";\n")
}
