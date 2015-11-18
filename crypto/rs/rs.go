package rs

import (
	"crypto/rand"
	"runtime"
)

var (
	Digital *RSGen
	Hex     *RSGen
	Base64  *RSGen
)

type RSGen struct {
	pipe chan []byte
}

func New(tab string) *RSGen {
	tl := byte(len(tab))
	if tl == 0 {
		panic("empty tab")
	}
	gen := &RSGen{pipe: make(chan []byte, runtime.NumCPU())}
	go func() {
		var (
			i int
			r []byte
			m []byte
		)
		for {
			r = make([]byte, 32)
			m = make([]byte, 32)
			rand.Read(r)
			for i = 0; i < 32; i++ {
				m[i] = tab[r[i]%tl]
			}
			gen.pipe <- m
		}
	}()
	return gen
}

func (gen *RSGen) Byte(len int) []byte {
	if len <= 0 {
		return nil
	} else if len <= 32 {
		return (<-gen.pipe)[:len]
	} else {
		bytes := make([]byte, len)
		for i := 0; i < len/32; i++ {
			copy(bytes[32*i:], <-gen.pipe)
		}
		if l := len % 32; l > 0 {
			copy(bytes[len-l:], (<-gen.pipe)[:l])
		}
		return bytes
	}
}

func (gen *RSGen) String(len int) string {
	return string(gen.Byte(len))
}

func init() {
	Digital = New("0123456789")
	Hex = New("0123456789abcdef")
	Base64 = New("./0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
}
