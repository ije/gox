package rs

import "crypto/rand"

var (
	Digital *RS
	Hex     *RS
	Base64  *RS
)

type RS struct {
	tab string
}

func New(tab string) *RS {
	if tab == "" {
		return &RS{"0123456789abcdef"}
	}
	return &RS{tab}
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
