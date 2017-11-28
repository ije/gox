/*

Log Package.

	package main

	import "github.com/ije/gox/log"

	func main() {
	    log, err := log.New("file:/tmp/go_test.log?buffer=64kb")
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

	"github.com/ije/gox/strconv"
)

var ErrorArgumentFormat = fmt.Errorf("invalid argument format")

type Logger struct {
	lock          sync.Mutex
	level         Level
	quite         bool
	buffer        []byte
	bufcap        int
	buflen        int
	prefix        string
	output        io.Writer
	flushTimer    *time.Timer
	logListeners  map[Level][]func(message []byte)
	errorListener func(data []byte, err error)
}

func New(url string) (*Logger, error) {
	l := &Logger{}
	return l, l.parseURL(url)
}

func (l *Logger) parseURL(url string) (err error) {
	if len(url) == 0 {
		return
	}

	i := strings.IndexByte(url, ':')
	if i == -1 {
		return fmt.Errorf("Incorrect URL '%s'", url)
	}
	driver, ok := drivers[strings.ToLower(url[:i])]
	if !ok {
		return fmt.Errorf("Unknown driver '%s'", url[:i])
	}
	url = url[i+1:]

	addr := url
	args := map[string]string{}
	if i = strings.IndexByte(url, '?'); i >= 0 {
		for _, q := range strings.Split(url[i+1:], "&") {
			kv := strings.SplitN(q, "=", 2)
			key := strings.TrimSpace(kv[0])
			value := ""
			if len(kv) == 2 {
				value = strings.TrimSpace(kv[1])
			}
			switch strings.ToLower(key) {
			case "prefix":
				l.SetPrefix(value)
			case "level":
				l.SetLevelByName(value)
			case "quite":
				l.SetQuite(value == "1" || strings.ToLower(value) == "true")
			case "buffer":
				bytes, err := strconv.ParseBytes(value)
				if err == nil {
					l.SetBuffer(int(bytes))
				}
			default:
				args[key] = value
			}
		}
		addr = url[:i]
	}

	wr, err := driver.Open(addr, args)
	if err == nil {
		l.SetOutput(wr)
	}
	return
}

func (l *Logger) SetLevel(level Level) {
	if level < L_DEBUG || level > L_FATAL {
		return
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	l.level = level
}

func (l *Logger) SetLevelByName(name string) {
	l.SetLevel(LevelByName(name))
}

func (l *Logger) SetPrefix(prefix string) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.prefix = strings.TrimSpace(prefix)
}

func (l *Logger) SetBuffer(cap int) {
	if cap < 1024 {
		return
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	if l.buffer == nil {
		l.bufcap = cap
		l.buffer = make([]byte, cap)
	} else if cap > l.bufcap {
		buffer := make([]byte, cap)
		copy(buffer, l.buffer)
		l.bufcap = cap
		l.buffer = buffer
	}
}

func (l *Logger) SetOutput(output io.Writer) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.output = output
}

func (l *Logger) SetQuite(quite bool) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.quite = quite
}

func (l *Logger) On(level Level, callback func(message []byte)) {
	if level < L_DEBUG || level > L_FATAL || callback == nil {
		return
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	if l.logListeners == nil {
		l.logListeners = map[Level][]func([]byte){level: []func([]byte){callback}}
	} else {
		l.logListeners[level] = append(l.logListeners[level], callback)
	}
}

func (l *Logger) OnWriteError(callback func(data []byte, err error)) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.errorListener = callback
}

func (l *Logger) Print(v ...interface{}) {
	l.log(-1, "", v...)
}

func (l *Logger) Printf(format string, v ...interface{}) {
	l.log(-1, format, v...)
}

func (l *Logger) Log(level string, v ...interface{}) {
	l.log(LevelByName(level), "", v...)
}

func (l *Logger) Logf(level, format string, v ...interface{}) {
	l.log(LevelByName(level), format, v...)
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

	// log event callback
	if callbacks, ok := l.logListeners[level]; ok {
		for _, callback := range callbacks {
			callback(buf)
		}
	}

	if !l.quite {
		if level < L_ERROR {
			os.Stdout.Write(buf)
		} else {
			os.Stderr.Write(buf)
		}
	}

	l.Write(buf)
}

func (l *Logger) fatal(format string, v ...interface{}) {
	l.log(L_FATAL, format, v...)
	l.FlushBuffer()
	os.Exit(1)
}

func (l *Logger) Write(p []byte) (n int, err error) {
	if n = len(p); n == 0 {
		return
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	if l.output == nil {
		return
	}

	defer func() {
		if err != nil && l.errorListener != nil {
			l.errorListener(p, err)
		}
	}()

	if l.bufcap > 0 {
		if n > l.bufcap/2 {
			n, err = l.output.Write(p)
			return
		}

		// Flush buffer
		if l.buflen > 0 && l.buflen+n > l.bufcap {
			_, err = l.output.Write(l.buffer[:l.buflen])
			if err != nil {
				return
			}
			l.buflen = 0
		}

		l.buflen += copy(l.buffer[l.buflen:], p)
		l.flushTime()
	} else {
		n, err = l.output.Write(p)
	}

	return
}

func (l *Logger) FlushBuffer() (n int, err error) {
	l.lock.Lock()
	defer l.lock.Unlock()

	if l.buflen == 0 || l.output == nil {
		n = l.buflen
		l.buflen = 0
		return
	}

	_, err = l.output.Write(l.buffer[:l.buflen])
	if err != nil {
		return
	}

	n = l.buflen
	l.buflen = 0
	return
}

func (l *Logger) flushTime() {
	if l.flushTimer != nil {
		l.flushTimer.Stop()
	}
	l.flushTimer = time.AfterFunc(5*time.Minute, func() {
		l.flushTimer = nil
		_, err := l.FlushBuffer()
		if err != nil {
			l.flushTime()
		}
	})
}
