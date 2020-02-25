package tunnel

import (
	logger "github.com/ije/gox/log"
)

var log = &logger.Logger{}

func init() {
	log.SetLevel(logger.L_INFO)
}

func SetLogger(l *logger.Logger) {
	if l != nil {
		log = l
	}
}

func SetLogLevel(level string) {
	log.SetLevelByName(level)
}
