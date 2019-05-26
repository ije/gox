package rs

import (
	"crypto/rand"
	"strings"
)

var (
	Digital *RS
	Hex     *RS
	Base64  *RS
)

type RS struct {
	tab string
}

func New(tab ...string) *RS {
	if len(tab) == 0 {
		return &RS{"0123456789abcdef"}
	}
	m := map[rune]struct{}{}
	for _, r := range strings.Join(tab, "") {
		m[r] = struct{}{}
	}
	runes := make([]rune, len(m))
	i := 0
	for r := range m {
		runes[i] = r
		i++
	}
	return &RS{string(runes)}
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

func init() {
	Digital = New("0123456789")
	Hex = New("0123456789abcdef")
	Base64 = New("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_")
}
