package log

import (
	"strings"
)

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

type Level int8

func LevelByName(name string) (l Level) {
	switch strings.ToLower(name) {
	case "debug":
		l = DEBUG
	case "info":
		l = INFO
	case "warn":
		l = WARN
	case "error":
		l = ERROR
	case "fatal":
		l = FATAL
	default:
		l = -1
	}
	return
}
