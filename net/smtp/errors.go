package smtp

import (
	"errors"
	"net/mail"
	"strings"
)

var (
	ErrEmptySender     = errors.New("Empty Sender")
	ErrEmptyRecipients = errors.New("Empty Recipients")
	ErrEmptySubject    = errors.New("Empty Subject")
	ErrEmptyContent    = errors.New("Empty Content")
)

type OTOSendError struct {
	Errors []*SendError
}

func (err *OTOSendError) Error() string {
	var es []string
	for _, e := range err.Errors {
		es = append(es, e.Error())
	}
	return strings.Join(es, ";\n")
}

type SendError struct {
	Message string
	From    *mail.Address
	To      AddressList
}

func (err *SendError) Error() string {
	return err.Message
}
