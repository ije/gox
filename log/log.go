/*
Package log implements a simple logging package.

	package main

	import (
		"github.com/ije/gox/log"
	)

	func main() {
	    l, err := log.New("file:/var/log/error.log?buffer=32kb")
	    if err != nil {
	        return
		}

	    l.Info("Hello World!")
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

type Logger struct {
	lock       sync.Mutex
	level      Level
	prefix     string
	output     io.Writer
	quite      bool
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
	url = strings.ReplaceAll(url, " ", "")
	if url == "" {
		return nil
	}

	fsn, path := utils.SplitByFirstByte(url, ':')
	fs, ok := registeredFileSystems[strings.ToLower(fsn)]
	if !ok {
		return fmt.Errorf("unknown fs protocol '%s'", fsn)
	}

	// handle format like file:///var/log/error.log
	if strings.HasPrefix(path, "//") {
		path = strings.TrimPrefix(path, "//")
	}
	if !strings.HasPrefix(path, "/") {
		path = "./" + path
	}

	args := map[string]string{}
	addr, query := utils.SplitByFirstByte(path, '?')
	for _, q := range strings.Split(query, "&") {
		key, value := utils.SplitByFirstByte(q, '=')
		if len(key) > 0 {
			switch strings.ToLower(key) {
			case "prefix":
				l.SetPrefix(value)
			case "level":
				l.SetLevelByName(value)
			case "quite":
				l.SetQuite(value == "" || value == "1" || strings.ToLower(value) == "true")
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

	output, err := fs.Open(addr, args)
	if err != nil {
		return
	}

	l.SetOutput(output)
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

func (l *Logger) Print(v ...interface{}) {
	l.log(-1, fmt.Sprint(v...))
}

func (l *Logger) Println(v ...interface{}) {
	l.log(-1, fmt.Sprintln(v...))
}

func (l *Logger) Printf(format string, v ...interface{}) {
	l.log(-1, fmt.Sprintf(fmt.Sprintf(format, v...)))
}

func (l *Logger) Debug(v ...interface{}) {
	l.log(L_DEBUG, fmt.Sprint(v...))
}

func (l *Logger) Debugln(v ...interface{}) {
	l.log(L_DEBUG, fmt.Sprintln(v...))
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	l.log(L_DEBUG, fmt.Sprintf(format, v...))
}

func (l *Logger) Info(v ...interface{}) {
	l.log(L_INFO, fmt.Sprint(v...))
}

func (l *Logger) Infoln(v ...interface{}) {
	l.log(L_INFO, fmt.Sprintln(v...))
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.log(L_INFO, fmt.Sprintf(format, v...))
}

func (l *Logger) Warn(v ...interface{}) {
	l.log(L_WARN, fmt.Sprint(v...))
}

func (l *Logger) Warnln(v ...interface{}) {
	l.log(L_WARN, fmt.Sprintln(v...))
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	l.log(L_WARN, fmt.Sprintf(format, v...))
}

func (l *Logger) Error(v ...interface{}) {
	l.log(L_ERROR, fmt.Sprint(v...))
}

func (l *Logger) Errorln(v ...interface{}) {
	l.log(L_ERROR, fmt.Sprintln(v...))
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.log(L_ERROR, fmt.Sprintf(format, v...))
}

func (l *Logger) Fatal(v ...interface{}) {
	l.fatal(fmt.Sprint(v...))
}

func (l *Logger) Fatalln(v ...interface{}) {
	l.fatal(fmt.Sprintln(v...))
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.fatal(fmt.Sprintf(format, v...))
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

func (l *Logger) log(level Level, msg string) {
	if level >= L_DEBUG && level < l.level {
		return
	}

	prefix := ""
	if level >= L_DEBUG && level <= L_FATAL {
		prefix = fmt.Sprintf("[%s] ", level)
	}
	if _prefix := l.prefix; len(_prefix) > 0 {
		prefix += _prefix + " "
	}

	bl := 20 + len(prefix) + len(msg) + 1
	buf := make([]byte, bl)
	now := time.Now()
	year, month, day := now.Date()
	hour, min, sec := now.Clock()
	i := pad(buf, 0, year, 4, '/')
	i = pad(buf, i, int(month), 2, '/')
	i = pad(buf, i, day, 2, ' ')
	i = pad(buf, i, hour, 2, ':')
	i = pad(buf, i, min, 2, ':')
	i = pad(buf, i, sec, 2, ' ')
	copy(buf[i:], prefix)
	copy(buf[i+len(prefix):], msg)

	if buf[bl-2] == '\n' {
		buf = buf[:bl-1]
	} else {
		buf[bl-1] = '\n'
	}

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
	l.log(L_FATAL, fmt.Sprintf(format, v...))
	l.FlushBuffer()
	os.Exit(1)
}

func (l *Logger) write(p []byte) (err error) {
	n := len(p)
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

		if l.flushTimer != nil {
			l.flushTimer.Stop()
		}

		if n > l.bufcap {
			n, err = l.output.Write(p)
		} else {
			copy(l.buffer[l.buflen:], p)
			l.buflen += n
			l.flushTimer = time.AfterFunc(time.Minute, func() {
				l.flushTimer = nil
				l.FlushBuffer()
			})
		}
	} else {
		n, err = l.output.Write(p)
	}
	return
}

func pad(p []byte, i int, u int, w int, suffix byte) int {
	i += w
	for j := 1; w > 0; j++ {
		p[i-j] = byte(u%10) + '0'
		u /= 10
		w--
	}
	p[i] = suffix
	return i + 1
}
