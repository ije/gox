package log

import (
	"io"
	"os"
	"path"

	"github.com/ije/gox/strconv"
	"github.com/ije/gox/utils"
)

var fws = map[string]*fileWriter{}

type fileWriter struct {
	filePath    string
	maxBytes    int
	writedBytes int
}

func (fw *fileWriter) Write(p []byte) (n int, err error) {
	if fw.maxBytes > 0 && fw.writedBytes > fw.maxBytes {
		if err = os.Rename(fw.filePath, fixFilePath(fw.filePath, 0)); err != nil {
			return
		}
		fw.writedBytes = 0
	}
	file, err := os.OpenFile(fw.filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	n, err = file.Write(p)
	fw.writedBytes += n
	return
}

func fixFilePath(filePath string, i int) (path string) {
	path, ext := utils.SplitByLastByte(filePath, '.')
	if i > 0 {
		path, _ = utils.SplitByLastByte(path, '.')
		path += "." + strconv.Itoa(i)
	}
	path += "." + ext
	if _, err := os.Lstat(path); err == nil || os.IsExist(err) {
		return fixFilePath(path, i+1)
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
		fw.writedBytes = int(fi.Size())
	}
	fws[filePath] = fw
	return
}

type fileLoggerDriver struct{}

func (fwd *fileLoggerDriver) Open(addr string, args map[string]string) (io.Writer, error) {
	maxBytes := 0
	if s, ok := args["maxBytes"]; ok && len(s) > 0 {
		i, err := strconv.ParseBytes(s)
		if err != nil {
			return nil, err
		}
		maxBytes = int(i)
	}
	return getFW(utils.PathClean(addr, true), maxBytes)
}

func init() {
	Register("file", &fileLoggerDriver{})
}
