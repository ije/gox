package rand

import (
	"crypto/rand"
)

var (
	Digital = &Rand{[]byte("0123456789")}
	Hex     = &Rand{[]byte("0123456789abcdefabcdef")}
	Base64  = &Rand{[]byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_")}
)

type Rand struct {
	KeyTable []byte
}

func (r *Rand) Bytes(size int) []byte {
	if size <= 0 {
		return nil
	}

	tl := byte(len(r.KeyTable))
	b := make([]byte, size)
	mapping := make([]byte, size)
	rand.Read(b)
	for i := 0; i < size; i++ {
		mapping[i] = r.KeyTable[b[i]%tl]
	}
	return mapping
}

func (r *Rand) String(size int) string {
	return string(r.Bytes(size))
}
