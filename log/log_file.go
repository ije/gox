package log

import (
	"io"
	"os"
	"path"
	"strconv"

	strconv2 "github.com/ije/gox/strconv"
	"github.com/ije/gox/utils"
)

var fws = map[string]*fileWriter{}

type fileWriter struct {
	writed   int
	maxBytes int
	filePath string
}

func (fw *fileWriter) Write(p []byte) (n int, err error) {
	if fw.maxBytes > 0 && fw.writed > fw.maxBytes {
		if err = os.Rename(fw.filePath, fw.Rename(0)); err != nil {
			return
		}
		fw.writed = 0
	}
	file, err := os.OpenFile(fw.filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	n, err = file.Write(p)
	fw.writed += n
	return
}

func (fw *fileWriter) Rename(c int) (name string) {
	name, ext := utils.SplitByLastByte(fw.filePath, '.')
	if c > 0 {
		name += "." + strconv.Itoa(c)
	}
	name += "." + ext
	if _, err := os.Lstat(name); err == nil || os.IsExist(err) {
		return fw.Rename(c + 1)
	}
	return
}

func getFW(filePath string, maxBytes int) (fw *fileWriter, err error) {
	fw, ok := fws[filePath]
	if ok {
		if maxBytes > 0 {
			fw.maxBytes = maxBytes
		}
		return
	}

	dir := path.Dir(filePath)
	if dir != "" && dir != "." {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return
		}
	}

	fw = &fileWriter{filePath: filePath, maxBytes: maxBytes}
	if fi, err := os.Lstat(filePath); err == nil {
		fw.writed = int(fi.Size())
	}
	fws[filePath] = fw
	return
}

type fwDriver struct{}

func (fwd *fwDriver) Open(addr string, args map[string]string) (io.Writer, error) {
	maxBytes := 0
	if s, ok := args["maxBytes"]; ok && len(s) > 0 {
		i, err := strconv2.ParseBytes(s)
		if err != nil {
			return nil, err
		}
		maxBytes = int(i)
	}
	return getFW(utils.PathClean(addr, true), maxBytes)
}

func init() {
	Register("file", &fwDriver{})
}
