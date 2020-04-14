package log

import (
	"io"
	"strings"
)

var registeredFileSystems = map[string]FileSystem{}

type FileSystem interface {
	Open(path string, args map[string]string) (writer io.Writer, err error)
}

func RegisterFileSystem(name string, fs FileSystem) {
	if fs != nil {
		registeredFileSystems[strings.ToLower(name)] = fs
	}
}
