package filehash

import (
	"crypto/md5"
	"hash/crc32"
	"io"
	"os"
)

const (
	HexTable = "0123456789abcdef"
)

func Hash(path string) (hash string, err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	p := make([]byte, 40)
	md5 := md5.New()
	io.Copy(md5, file)
	for i, v := range md5.Sum(nil) {
		p[i*2] = HexTable[v>>4]
		p[i*2+1] = HexTable[v&0x0f]
	}
	crc32 := crc32.NewIEEE()
	file.Seek(0, io.SeekStart)
	io.Copy(crc32, file)
	crc := crc32.Sum32()
	for i := 1; i <= 8; i++ {
		p[40-i] = HexTable[crc%16]
		crc /= 16
	}

	hash = string(p)
	return
}
