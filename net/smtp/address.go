package smtp

import (
	"net/mail"
	"strings"

	"github.com/ije/gox/valid"
)

func Address(a ...string) *mail.Address {
	var name string
	var address string
	if len(a) == 1 {
		if len(a[0]) > 0 {
			addr, err := mail.ParseAddress(a[0])
			if err == nil {
				return addr
			}
		}
	} else if len(a) > 1 {
		name = a[0]
		address = a[1]
	}

	if len(address) == 0 || !valid.IsEmail(address) {
		return nil
	}

	return &mail.Address{
		Name:    name,
		Address: address,
	}
}

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
