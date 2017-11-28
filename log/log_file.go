package log

import (
	"io"
	"os"
	"path"
	"time"

	"github.com/ije/gox/strconv"
	"github.com/ije/gox/utils"
)

type fileWriter struct {
	filePath       string
	fileDateFormat string
	maxFileBytes   int64
	writedBytes    int64
}

func (w *fileWriter) Write(p []byte) (n int, err error) {
	if w.maxFileBytes > 0 && w.writedBytes > w.maxFileBytes {
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
	} else {
		return w.filePath
	}
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

func newWriter(filePath string, fileDateFormat string, maxFileBytes int64) (w *fileWriter, err error) {
	dir := path.Dir(filePath)
	if dir != "" && dir != "." {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return
		}
	}

	w = &fileWriter{filePath: filePath, fileDateFormat: fileDateFormat, maxFileBytes: maxFileBytes}
	if fi, err := os.Lstat(w.fixedFilePath()); err == nil {
		w.writedBytes = fi.Size()
	}
	return
}

type fileLoggerDriver struct{}

func (d *fileLoggerDriver) Open(addr string, args map[string]string) (io.Writer, error) {
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

	return newWriter(utils.CleanPath(addr), args["fileDateFormat"], maxFileBytes)
}

func init() {
	Register("file", &fileLoggerDriver{})
}
