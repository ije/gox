package smtp

import (
	"net/mail"
	"strings"
)

type SendError struct {
	Message string
	From    *mail.Address
	To      AddressList
}

func (err *SendError) Error() string {
	return err.Message
}

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
