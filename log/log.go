/*

Log Package.

	package main

	import "github.com/ije/go/log"

	func main() {
	    log, err := log.New("file:/var/log/go/test.log?buffer=64kb")
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

	"github.com/ije/go/utils"
)

type Logger struct {
	lock         sync.Mutex
	level        Level
	quite        bool
	bufcap       int
	buflen       int
	buffer       []byte
	writer       io.Writer
	timer        *time.Timer
	onWriteError func(data []byte, err error)
	logListeners map[Level][]func(message []byte)
}

func New(url string) (l *Logger, err error) {
	l = &Logger{}
	if err = l.parseURL(url); err != nil {
		l = nil
	}
	return
}

func (l *Logger) parseURL(url string) (err error) {
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
			case "level":
				l.SetLevelByName(value)
			case "quite":
				l.SetQuite(strings.ToLower(value) == "true" || value == "1")
			case "buffer":
				i, err := utils.ParseByte(value)
				if err == nil {
					l.SetBuffer(int(i))
				}
			default:
				args[key] = value
			}
		}
		addr = url[:i]
	}

	wr, err := driver.Open(addr, args)
	if err == nil {
		l.SetWriter(wr)
	}
	return
}

func (l *Logger) SetWriter(writer io.Writer) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.writer = writer
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

func (l *Logger) SetBuffer(maxMemory int) {
	if maxMemory < 1024 {
		return
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	if l.buffer == nil {
		l.bufcap = maxMemory
		l.buffer = make([]byte, maxMemory)
	} else if maxMemory > l.bufcap {
		buffer := make([]byte, maxMemory)
		copy(buffer, l.buffer)
		l.bufcap = maxMemory
		l.buffer = buffer
	}
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
		l.logListeners = map[Level][]func(message []byte){level: []func(message []byte){callback}}
	} else {
		l.logListeners[level] = append(l.logListeners[level], callback)
	}
}

func (l *Logger) OnWriteError(callback func(message []byte, err error)) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.onWriteError = callback
}

func (l *Logger) Flush() (n int, err error) {
	l.lock.Lock()
	defer l.lock.Unlock()

	if l.buflen > 0 {
		n, err = l.writer.Write(l.buffer[:l.buflen])
		if err != nil {
			return
		}
		l.buflen = 0
	}
	return
}

func (l *Logger) Print(v ...interface{}) {
	l.log(-1, "", v...)
}

func (l *Logger) Printf(format string, v ...interface{}) {
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
	l.log(L_FATAL, "", v...)
	l.Flush()
	os.Exit(1)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.log(L_FATAL, format, v...)
	l.Flush()
	os.Exit(1)
}

func (l *Logger) log(level Level, format string, v ...interface{}) {
	if level != -1 && level < l.level {
		return
	}

	var lp, msg string
	var s = 28
	switch level {
	case L_FATAL:
		lp = "[fatal] "
	case L_ERROR:
		lp = "[error] "
	case L_WARN:
		lp = "[warn] "
		s--
	case L_INFO:
		lp = "[info] "
		s--
	case L_DEBUG:
		lp = "[debug] "
	case -1:
		s = 20
	}

	if fl := len(format); fl > 0 {
		if format[fl-1] != '\n' {
			format += "\n"
		}
		msg = fmt.Sprintf(format, v...)
	} else {
		msg = fmt.Sprintln(v...)
	}

	var i int
	p := make([]byte, s+len(msg))
	iaa := func(u int, wid int, end byte) {
		i += wid
		for j := 1; wid > 0; j++ {
			p[i-j] = byte(u%10) + '0'
			u /= 10
			wid--
		}
		i++
		p[i-1] = end
	}
	now := time.Now()
	year, month, day := now.Date()
	hour, min, sec := now.Clock()
	iaa(year, 4, '/')
	iaa(int(month), 2, '/')
	iaa(day, 2, ' ')
	iaa(hour, 2, ':')
	iaa(min, 2, ':')
	iaa(sec, 2, ' ')
	copy(p[20:], lp)
	copy(p[s:], msg)

	// log event callback
	if callbacks, ok := l.logListeners[level]; ok {
		for _, callback := range callbacks {
			callback(p)
		}
	}

	if !l.quite {
		if level <= L_INFO {
			os.Stdout.Write(p)
		} else {
			os.Stderr.Write(p)
		}
	}

	l.Write(p)
}

func (l *Logger) Write(p []byte) (n int, err error) {
	if n = len(p); n == 0 {
		return
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	if l.writer == nil {
		return
	}

	defer func() {
		if err != nil && l.onWriteError != nil {
			l.onWriteError(p, err)
		}
	}()

	if l.bufcap > 0 {
		if l.buflen+n > l.bufcap { // Flush
			if l.buflen > 0 {
				n, err = l.writer.Write(l.buffer[:l.buflen])
				if err != nil {
					return
				}
				l.buflen = 0
			}
		}

		if n >= l.bufcap {
			n, err = l.writer.Write(p)
			return
		}

		l.buflen += copy(l.buffer[l.buflen:], p)

		if l.timer != nil {
			l.timer.Stop()
		}
		l.timer = time.AfterFunc(5*time.Minute, func() {
			l.Flush()
		})
	} else {
		n, err = l.writer.Write(p)
	}

	return
}
