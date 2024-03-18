package rs

import (
	"crypto/rand"
)

var (
	Digital = New("0123456789")
	Hex     = New("0123456789abcdef")
	Base64  = New("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_")
)

type RS struct {
	tab []byte
}

func New(tab ...string) *RS {
	if len(tab) == 0 {
		return &RS{[]byte("0123456789abcdef")}
	}
	set := map[byte]struct{}{}
	for _, s := range tab {
		for _, c := range []byte(s) {
			set[c] = struct{}{}
		}
	}
	bytes := make([]byte, len(set))
	i := 0
	for c := range set {
		bytes[i] = c
		i++
	}
	return &RS{bytes}
}

func (rs *RS) Bytes(size int) []byte {
	if size <= 0 {
		return nil
	}

	tl := byte(len(rs.tab))
	r := make([]byte, size)
	ret := make([]byte, size)
	rand.Read(r)
	for i := 0; i < size; i++ {
		ret[i] = rs.tab[r[i]%tl]
	}
	return ret
}

func (rs *RS) String(len int) string {
	return string(rs.Bytes(len))
}
