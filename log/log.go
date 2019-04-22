/*
Package log ...

	package main

	import (
		"sync"
		"github.com/ije/gox/log"
	)

	func main() {
	    log, err := log.New("file:/var/log/error.log?buffer=32kb")
	    if err != nil {
	        return
	    }
	    log.Info("Hello World!")
	}

*/
package log

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ije/gox/utils"
)

var ErrorArgumentFormat = fmt.Errorf("invalid argument format")

type Logger struct {
	lock       sync.Mutex
	level      Level
	quite      bool
	prefix     string
	output     io.Writer
	buffer     []byte
	bufcap     int
	buflen     int
	flushTimer *time.Timer
}

func New(url string) (*Logger, error) {
	l := &Logger{}
	return l, l.parseURL(url)
}

func (l *Logger) parseURL(url string) (err error) {
	if url == "" {
		err = fmt.Errorf("invalid url")
		return
	}

	path, query := utils.SplitByFirstByte(url, '?')
	name, addr := utils.SplitByFirstByte(path, ':')
	driver, ok := drivers[strings.ToLower(name)]
	if !ok {
		return fmt.Errorf("Unknown driver '%s'", name)
	}

	args := map[string]string{}
	for _, q := range strings.Split(query, "&") {
		key, value := utils.SplitByFirstByte(q, '=')
		if len(key) > 0 {
			switch strings.ToLower(key) {
			case "prefix":
				l.SetPrefix(value)
			case "level":
				l.SetLevelByName(value)
			case "quite":
				l.SetQuite(value == "1" || strings.ToLower(value) == "true")
			case "buffer":
				bytes, err := utils.ParseBytes(value)
				if err == nil {
					l.SetBuffer(int(bytes))
				}
			default:
				args[key] = value
			}
		}
	}

	wr, err := driver.Open(addr, args)
	if err == nil {
		l.SetOutput(wr)
	}
	return
}

func (l *Logger) SetLevel(level Level) {
	if level >= L_DEBUG && level <= L_FATAL {
		l.level = level
	}
}

func (l *Logger) SetLevelByName(name string) {
	l.SetLevel(LevelByName(name))
}

func (l *Logger) SetPrefix(prefix string) {
	l.prefix = strings.TrimSpace(prefix)
}

func (l *Logger) SetQuite(quite bool) {
	l.quite = quite
}

func (l *Logger) SetBuffer(cap int) {
	if cap < 32 {
		return
	}

	l.FlushBuffer()

	l.lock.Lock()
	defer l.lock.Unlock()

	l.buffer = make([]byte, cap)
	l.bufcap = cap
	l.buflen = 0
}

func (l *Logger) SetOutput(output io.Writer) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.output = output
}

func (l *Logger) Log(v ...interface{}) {
	l.log(-1, "", v...)
}

func (l *Logger) Logf(format string, v ...interface{}) {
	l.log(-1, format, v...)
}

func (l *Logger) Debug(v ...interface{}) {
	l.log(L_DEBUG, "", v...)
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	l.log(L_DEBUG, format, v...)
}

func (l *Logger) Info(v ...interface{}) {
	l.log(L_INFO, "", v...)
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.log(L_INFO, format, v...)
}

func (l *Logger) Warn(v ...interface{}) {
	l.log(L_WARN, "", v...)
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	l.log(L_WARN, format, v...)
}

func (l *Logger) Error(v ...interface{}) {
	l.log(L_ERROR, "", v...)
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.log(L_ERROR, format, v...)
}

func (l *Logger) Fatal(v ...interface{}) {
	l.fatal("", v...)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.fatal(format, v...)
}

func (l *Logger) FlushBuffer() (err error) {
	l.lock.Lock()
	defer l.lock.Unlock()

	if l.buflen > 0 {
		if l.output != nil {
			_, err = l.output.Write(l.buffer[:l.buflen])
			if err != nil {
				return
			}
		}
		l.buflen = 0
	}
	return
}

func (l *Logger) log(level Level, format string, v ...interface{}) {
	if level != -1 && level < l.level {
		return
	}

	var prefix, msg string
	switch level {
	case L_FATAL:
		prefix = "[fatal] "
	case L_ERROR:
		prefix = "[error] "
	case L_WARN:
		prefix = "[warn] "
	case L_INFO:
		prefix = "[info] "
	case L_DEBUG:
		prefix = "[debug] "
	}
	if len(l.prefix) > 0 {
		prefix += l.prefix + " "
	}

	if l := len(format); l > 0 {
		if format[l-1] != '\n' {
			format += "\n"
		}
		msg = fmt.Sprintf(format, v...)
	} else {
		msg = fmt.Sprintln(v...)
	}

	var i int
	buf := make([]byte, 20+len(prefix)+len(msg))
	fd := func(u int, w int, suffix byte) {
		i += w
		for j := 1; w > 0; j++ {
			buf[i-j] = byte(u%10) + '0'
			u /= 10
			w--
		}
		i++
		buf[i-1] = suffix
	}

	now := time.Now()
	year, month, day := now.Date()
	hour, min, sec := now.Clock()
	fd(year, 4, '/')
	fd(int(month), 2, '/')
	fd(day, 2, ' ')
	fd(hour, 2, ':')
	fd(min, 2, ':')
	fd(sec, 2, ' ')
	copy(buf[20:], prefix)
	copy(buf[20+len(prefix):], msg)

	if !l.quite {
		l.lock.Lock()
		if level < L_ERROR {
			os.Stdout.Write(buf)
		} else {
			os.Stderr.Write(buf)
		}
		l.lock.Unlock()
	}

	l.write(buf)
}

func (l *Logger) fatal(format string, v ...interface{}) {
	l.log(L_FATAL, format, v...)
	l.FlushBuffer()
	os.Exit(1)
}

func (l *Logger) write(p []byte) (n int, err error) {
	n = len(p)
	if n == 0 {
		return
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	if l.output == nil {
		return
	}

	if l.bufcap > 0 {
		// flush buffer
		if l.buflen > 0 && l.buflen+n > l.bufcap {
			_, err = l.output.Write(l.buffer[:l.buflen])
			if err != nil {
				return
			}
			l.buflen = 0
		}

		if n > l.bufcap {
			n, err = l.output.Write(p)
			return
		}

		copy(l.buffer[l.buflen:], p)
		l.buflen += n

		if l.flushTimer != nil {
			l.flushTimer.Stop()
		}
		l.flushTimer = time.AfterFunc(time.Minute, func() {
			l.flushTimer = nil
			l.FlushBuffer()
		})
	} else {
		n, err = l.output.Write(p)
	}
	return
}
