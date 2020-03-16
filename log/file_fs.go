package log

import (
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"time"

	"gox/utils"
)

type fileWriter struct {
	filePath       string
	fileDateFormat string
	maxFileSize    int64
	writedBytes    int64
}

func (w *fileWriter) Write(p []byte) (n int, err error) {
	if w.maxFileSize > 0 && w.writedBytes > w.maxFileSize {
		if err = os.Rename(w.filePath, appendFileIndex(w.fixedFilePath(), 0)); err != nil {
			return
		}
		w.writedBytes = 0
	}
	file, err := os.OpenFile(w.fixedFilePath(), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	n, err = file.Write(p)
	w.writedBytes += int64(n)
	return
}

func (w *fileWriter) fixedFilePath() (path string) {
	if len(w.fileDateFormat) > 0 {
		name, ext := utils.SplitByLastByte(w.filePath, '.')
		return name + "-" + time.Now().Format(w.fileDateFormat) + "." + ext
	}
	return w.filePath
}

func appendFileIndex(path string, i int) string {
	name, ext := utils.SplitByLastByte(path, '.')
	if i > 0 {
		name, _ = utils.SplitByLastByte(name, '_')
		name += "_" + strconv.Itoa(i)
	}

	path = name + "." + ext
	if _, err := os.Lstat(path); err == nil || os.IsExist(err) {
		return appendFileIndex(path, i+1)
	}

	return path
}

func newWriter(filePath string, fileDateFormat string, maxFileSize int64) (w *fileWriter, err error) {
	dir := path.Dir(filePath)
	fi, err := os.Stat(dir)
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
	} else if err == nil && !fi.IsDir() {
		err = fmt.Errorf("invalid filePath %s", filePath)
	}
	if err != nil {
		return
	}

	w = &fileWriter{filePath: filePath, fileDateFormat: fileDateFormat, maxFileSize: maxFileSize}
	if fi, err := os.Lstat(w.fixedFilePath()); err == nil {
		w.writedBytes = fi.Size()
	}
	return
}

type fileFS struct{}

func (d *fileFS) Open(path string, args map[string]string) (io.Writer, error) {
	var fileDateFormat string
	var maxFileSize int64

	if val, ok := args["fileDateFormat"]; ok {
		if val == "" {
			val = "2006-01-02"
		}
		fileDateFormat = val
	}

	if val, ok := args["maxFileSize"]; ok && len(val) > 0 {
		i, err := utils.ParseBytes(val)
		if err != nil {
			return nil, fmt.Errorf("invalid maxFileSize argument")
		}
		maxFileSize = i
	}

	return newWriter(path, fileDateFormat, maxFileSize)
}

func init() {
	Register("file", &fileFS{})
}
