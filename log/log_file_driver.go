package log

import (
	"io"
	"os"

	"github.com/ije/go/utils"
)

type fwDriver struct{}

func (fwd *fwDriver) Open(addr string, args map[string]string) (io.Writer, error) {
	maxBytes := 0
	if s, ok := args["maxBytes"]; ok && len(s) > 0 {
		i, err := utils.ParseByte(s)
		if err != nil {
			return nil, err
		}
		maxBytes = int(i)
	}
	return getFW(utils.PathClean(addr), maxBytes)
}

var fws = map[string]*fileWriter{}

func getFW(filepath string, maxBytes int) (fw *fileWriter, err error) {
	fw, ok := fws[filepath]
	if ok {
		if maxBytes > 0 {
			fw.maxBytes = maxBytes
		}
		return
	}

	var dir = "."
	for i := len(filepath) - 1; i > 0; i-- {
		if filepath[i] == '/' || filepath[i] == '\\' {
			dir = filepath[:i]
			break
		}
	}
	if dir != "" && dir != "." {
		if err = os.MkdirAll(dir, 0744); err != nil {
			return
		}
	}

	fw = &fileWriter{filepath: filepath, maxBytes: maxBytes}
	if fi, err := os.Lstat(filepath); err == nil {
		fw.writed = int(fi.Size())
	}
	fws[filepath] = fw
	return
}

func init() {
	Register("file", &fwDriver{})
}
