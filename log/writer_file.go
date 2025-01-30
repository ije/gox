package log

import (
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/ije/gox/utils"
)

type fileWriter struct {
	fileName       string
	fileDateFormat string
	maxFileSize    int64
	writedBytes    int64
}

func (w *fileWriter) Write(p []byte) (n int, err error) {
	if w.maxFileSize > 0 && w.writedBytes > w.maxFileSize {
		if err = os.Rename(w.fileName, appendFileIndex(w.fixedFilePath(), 0)); err != nil {
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
		name, ext := utils.Split2Last(w.fileName, '.')
		return name + "-" + time.Now().Format(w.fileDateFormat) + "." + ext
	}
	return w.fileName
}

func appendFileIndex(path string, i int) string {
	name, ext := utils.Split2Last(path, '.')
	if i > 0 {
		name, _ = utils.Split2Last(name, '_')
		name += "_" + strconv.Itoa(i)
	}

	path = name + "." + ext
	if _, err := os.Lstat(path); err == nil || os.IsExist(err) {
		return appendFileIndex(path, i+1)
	}

	return path
}

func newFileWriter(fileName string, fileDateFormat string, maxFileSize int64) (w *fileWriter, err error) {
	dir := path.Dir(fileName)
	fi, err := os.Stat(dir)
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
	} else if err == nil && !fi.IsDir() {
		err = fmt.Errorf("invalid filePath %s", fileName)
	}
	if err != nil {
		return
	}

	w = &fileWriter{fileName: fileName, fileDateFormat: fileDateFormat, maxFileSize: maxFileSize}
	if fi, err := os.Lstat(w.fixedFilePath()); err == nil {
		w.writedBytes = fi.Size()
	}
	return
}

type fWriter struct{}

func (d *fWriter) Open(path string, args map[string]string) (io.Writer, error) {
	var maxFileSize int64
	var fileDateFormat string

	if val, ok := args["maxFileSize"]; ok && len(val) > 0 {
		i, err := utils.ParseBytes(val)
		if err != nil {
			return nil, fmt.Errorf("invalid maxFileSize argument")
		}
		maxFileSize = i
	}

	if val, ok := args["fileDateFormat"]; ok {
		fileDateFormat = val
	}

	return newFileWriter(path, fileDateFormat, maxFileSize)
}

func init() {
	RegisterLogWriter("file", &fWriter{})
}
