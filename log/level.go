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
