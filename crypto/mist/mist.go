package mist

import (
	"crypto/rand"
	"runtime"
)

var (
	Hex    *Mist
	Base64 *Mist
)

type Mist struct {
	pipe chan []byte
}

func New(tab string) *Mist {
	tl := byte(len(tab))
	mist := &Mist{pipe: make(chan []byte, runtime.NumCPU())}
	go func() {
		var (
			i int
			r []byte
			m []byte
		)
		for {
			r = make([]byte, 32)
			rand.Read(r)
			m = make([]byte, 32)
			for i = 0; i < 32; i++ {
				m[i] = tab[r[i]%tl]
			}
			mist.pipe <- m
		}
	}()
	return mist
}

func (mist *Mist) Byte(len int) []byte {
	switch {
	case len <= 0:
		return nil
	case len <= 32:
		return (<-mist.pipe)[:len]
	default:
		ms := make([]byte, len)
		for i := 0; i < len/32; i++ {
			copy(ms[32*i:], <-mist.pipe)
		}
		if l := len % 32; l > 0 {
			copy(ms[len-l:], (<-mist.pipe)[:l])
		}
		return ms
	}
}

func (mist *Mist) String(len int) string {
	return string(mist.Byte(len))
}

func init() {
	Hex = New("0123456789abcdef")
	Base64 = New("./0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
}
