package log

import (
	"os"
	"strconv"
	"time"
)

type fileWriter struct {
	writed     int
	maxBytes   int
	filepath   string
	splitByDay bool
}

func (fw *fileWriter) Write(p []byte) (n int, err error) {
	if fw.splitByDay {
		// todo: split by day
	} else if fw.maxBytes > 0 && fw.writed > fw.maxBytes {
		if err = os.Rename(fw.filepath, fw.Rename(0)); err != nil {
			return
		}
		fw.writed = 0
	}
	file, err := os.OpenFile(fw.filepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	n, err = file.Write(p)
	fw.writed += n
	return
}

func (fw *fileWriter) Rename(c int) (name string) {
	filename, ext := fw.filepath, ""
	for i := len(filename) - 1; i > 0; i-- {
		if filename[i] == '.' {
			ext = filename[i:]
			filename = filename[:i]
			break
		}
	}
	name = filename + "." + time.Now().Format("060102")
	if c > 0 {
		name += "-" + strconv.Itoa(c)
	}
	name += ext
	if _, err := os.Lstat(name); err == nil {
		return fw.Rename(c + 1)
	}
	return
}
