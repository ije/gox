package smtp

import (
	"net/mail"
	"strings"
)

type AddressList []*mail.Address

func (list AddressList) String() string {
	var ss []string
	for _, addr := range list {
		ss = append(ss, addr.String())
	}
	return strings.Join(ss, ", ")
}

func (list AddressList) List() []string {
	var ss []string
	for _, addr := range list {
		ss = append(ss, addr.Address)
	}
	return ss
}
