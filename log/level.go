package log

import (
	"strings"
)

const (
	L_DEBUG Level = iota
	L_INFO
	L_WARN
	L_ERROR
	L_FATAL
)

type Level int8

func (l Level) String() string {
	var name string
	switch l {
	case L_FATAL:
		name = "fatal"
	case L_ERROR:
		name = "error"
	case L_WARN:
		name = "warn"
	case L_INFO:
		name = "info"
	case L_DEBUG:
		name = "debug"
	}
	return name
}

func LevelByName(name string) (l Level) {
	switch strings.ToLower(name) {
	case "debug":
		l = L_DEBUG
	case "info":
		l = L_INFO
	case "warn":
		l = L_WARN
	case "error":
		l = L_ERROR
	case "fatal":
		l = L_FATAL
	default:
		l = -1
	}
	return
}
