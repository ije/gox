package log

import (
	"io"
	"strings"
)

var registry = map[string]LogWriter{}

type LogWriter interface {
	Open(path string, args map[string]string) (writer io.Writer, err error)
}

func RegisterLogWriter(name string, fs LogWriter) {
	if fs != nil {
		registry[strings.ToLower(name)] = fs
	}
}
