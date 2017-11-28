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
	filePath     string
	maxFileBytes int64
	writedBytes  int64
}

func (fw *fileWriter) Write(p []byte) (n int, err error) {
	if fw.maxFileBytes > 0 && fw.writedBytes > fw.maxFileBytes {
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
	fw.writedBytes += int64(n)
	return
}

func fixFilePath(filePath string, i int) (path string) {
	path, ext := utils.SplitByLastByte(filePath, '.')
	if i > 0 {
		path, _ = utils.SplitByLastByte(path, '.')
		path += "_" + strconv.Itoa(i)
	}
	path += "." + ext
	if _, err := os.Lstat(path); err == nil || os.IsExist(err) {
		return fixFilePath(path, i+1)
	}
	return
}

func getFW(filePath string, maxFileBytes int64) (fw *fileWriter, err error) {
	fw, ok := fws[filePath]
	if ok {
		if maxFileBytes > 0 {
			fw.maxFileBytes = maxFileBytes
		}
		return
	}

	dir := path.Dir(filePath)
	if dir != "" && dir != "." {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return
		}
	}

	fw = &fileWriter{filePath: filePath, maxFileBytes: maxFileBytes}
	if fi, err := os.Lstat(filePath); err == nil {
		fw.writedBytes = fi.Size()
	}
	fws[filePath] = fw
	return
}

type fileLoggerDriver struct{}

func (fwd *fileLoggerDriver) Open(addr string, args map[string]string) (io.Writer, error) {
	var maxFileBytes int64

	val, ok := args["maxFileBytes"]
	if !ok {
		val, ok = args["maxBytes"]
	}
	if ok && len(val) > 0 {
		i, err := strconv.ParseBytes(val)
		if err != nil {
			return nil, ErrorArgumentFormat
		}
		maxFileBytes = i
	}

	return getFW(utils.CleanPath(addr), maxFileBytes)
}

func init() {
	Register("file", &fileLoggerDriver{})
}
