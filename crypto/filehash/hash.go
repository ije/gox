package filehash

import (
	"crypto/md5"
	"hash/crc32"
	"io"
	"os"
)

const (
	hexTable = "0123456789abcdef"
)

func HashFile(file string) (hash string, err error) {
	f, err := os.Open(file)
	if err != nil {
		return
	}
	defer f.Close()

	return Hash(f)
}

func Hash(file io.ReadSeeker) (hash string, err error) {
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return
	}

	p := make([]byte, 40)
	md5 := md5.New()
	io.Copy(md5, file)
	for i, v := range md5.Sum(nil) {
		p[i*2] = hexTable[v>>4]
		p[i*2+1] = hexTable[v&0x0f]
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return
	}

	crc32 := crc32.NewIEEE()
	io.Copy(crc32, file)
	crc := crc32.Sum32()
	for i := 1; i <= 8; i++ {
		p[40-i] = hexTable[crc%16]
		crc /= 16
	}

	hash = string(p)
	return
}
