package utils

import (
	"encoding/binary"
	"encoding/json"
	"html"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func CatchExit(callback func()) {
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Kill, os.Interrupt, syscall.SIGTERM)

	go func() {
		for {
			switch <-sig {
			case os.Kill, os.Interrupt, syscall.SIGTERM:
				callback()
				os.Exit(1)
			}
		}
	}()
}

func CopyFile(src, dst string) (int64, error) {
	if src == dst {
		return 0, nil
	}
	sf, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	if _, err := os.Lstat(dst); err != nil && !os.IsNotExist(err) {
		return 0, err
	}
	df, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer df.Close()
	return io.Copy(df, sf)
}

func JSONUnmarshalFile(filename string, v interface{}) (err error) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(v)
}

func JSONMarshalFile(filename string, v interface{}) (err error) {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(v)
}

func HtmlToText(s string, limit int, unescape bool) string {
	var i int
	var tagStart bool
	var sigStart bool
	var sig []rune
	var runes = make([]rune, len(s))
	var add = func(c rune) {
		runes[i] = c
		i++
	}
	for _, r := range []rune(s) {
		if limit > -1 && i >= limit {
			break
		}
		switch r {
		case '<':
			tagStart = true
		case '>':
			tagStart = false
		case '&':
			if tagStart {
				break
			}
			sigStart = true
			sig = nil
		case ';':
			if tagStart {
				break
			}
			if sigStart {
				if unescape {
					if tr := []rune(html.UnescapeString(string(append([]rune{'&'}, append(sig, ';')...)))); len(tr) > 0 {
						add(tr[0])
					}
				} else {
					for _, r := range "&" + string(sig) + ";" {
						add(r)
					}
				}
			}
			sigStart = false
			sig = nil
		default:
			if tagStart {
				break
			}
			if sigStart {
				if r >= 'A' && r <= 'Z' {
					r += 32 // ToLower
				}
				sig = append(sig, r)
				break
			}
			add(r)
		}
	}
	return string(runes[:i])
}

func SplitToLines(s string, chars string) (lines []string) {
	for i, j, l := 0, 0, len(s); i < l; i++ {
		switch s[i] {
		case '\r', '\n':
			if i > j {
				lines = append(lines, s[j:i])
			}
			j = i + 1
		default:
			if i == l-1 && j < l {
				lines = append(lines, s[j:])
			}
		}
	}
	return
}

func SplitByLastByte(s string, c byte) (string, string) {
	for i := len(s) - 1; i > 0; i-- {
		if s[i] == c {
			return s[:i], s[i+1:]
		}
	}
	return s, ""
}

func FileExt(filename string) (ext string) {
	for i := len(filename) - 1; i > 0; i-- {
		if filename[i] == '.' {
			ext = filename[i+1:]
			break
		}
	}
	return
}

// PathClean has the same function with path.Clean(strings.ToLower(strings.Replace(strings.TrimSpace(s), "\\", "/", -1))),
// but it's faster!
func PathClean(path string, toLower bool) string {
	pl := len(path)
	if pl == 0 {
		return "."
	}
	var n int
	var c byte
	var root bool
	var newpath = make([]byte, pl)
	for i := 0; i < pl; i++ {
		switch c = path[i]; c {
		case ' ':
			if n == 0 {
				continue
			}
			newpath[n] = ' '
			n++
		case '/', '\\':
			if n > 0 {
				if newpath[n-1] == '/' {
					continue
				} else if newpath[n-1] == '.' && n > 1 && newpath[n-2] == '/' {
					n--
					continue
				}
			}
			if n == 0 {
				root = true
			}
			newpath[n] = '/'
			n++
		case '.':
			if n > 1 && newpath[n-1] == '.' && newpath[n-2] == '/' {
				if n = n - 2; n > 0 {
					for n = n - 1; n > 0; n-- {
						if newpath[n] == '/' {
							break
						}
					}
				}
				continue
			}
			newpath[n] = '.'
			n++
		default:
			if toLower && c >= 'A' && c <= 'Z' {
				c += 32 // ToLower
			}
			newpath[n] = c
			n++
		}
	}
	// trim right spaces
	if n > 0 && newpath[n-1] == ' ' {
		for n > 0 && newpath[n-1] == ' ' {
			n--
		}
	}
	if n > 1 && newpath[n-1] == '.' && newpath[n-2] == '/' {
		n--
	}
	if n > 0 && newpath[n-1] == '/' && (!root || n > 1) {
		n--
	}
	if n == 0 {
		return "."
	}
	return string(newpath[:n])
}

func Ipv4ToLong(ipStr string) uint32 {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return 0
	}
	ip = ip.To4()
	return binary.BigEndian.Uint32(ip)
}

func LongToIpv4(ipLong uint32) string {
	ipByte := make([]byte, 4)
	binary.BigEndian.PutUint32(ipByte, ipLong)
	ip := net.IP(ipByte)
	return ip.String()
}
