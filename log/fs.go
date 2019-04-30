package log

import (
	"io"
	"strings"
)

var fss = map[string]FS{}

type FS interface {
	Open(path string, args map[string]string) (writer io.Writer, err error)
}

func Register(name string, fs FS) {
	if fs != nil {
		fss[strings.ToLower(name)] = fs
	}
}
