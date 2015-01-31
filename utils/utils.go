package utils

import (
	"encoding/binary"
	"encoding/json"
	"html"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"syscall"
)

const (
	B int64 = 1 << (10 * iota)
	KB
	MB
	GB
	TB
	PB
)

// Go is a basic promise implementation: it wraps calls a function in a goroutine,
// and returns a channel which will later return the function's return value.
func Go(f func() error) chan error {
	ch := make(chan error, 1)
	go func() {
		ch <- f()
	}()
	return ch
}

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

func ParseByte(s string) (int64, error) {
	if sl := len(s); sl > 0 {
		b := B
	BeginParse:
		switch sl--; s[sl] {
		case 'b', 'B':
			if sl == len(s)-1 {
				goto BeginParse
			}
		case 'k', 'K':
			b = KB
		case 'm', 'M':
			b = MB
		case 'g', 'G':
			b = GB
		case 't', 'T':
			b = TB
		case 'p', 'P':
			b = PB
		default:
			sl++
		}
		if sl == 0 {
			return 0, strconv.ErrSyntax
		}
		i, err := strconv.ParseInt(s[:sl], 10, 64)
		if err != nil {
			return 0, strconv.ErrSyntax
		}
		b *= i
		return b, nil
	}
	return 0, strconv.ErrSyntax
}

func JSONUnmarshalFile(filename string, v interface{}) (err error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(regexp.MustCompile(`,\s*(\]|\})`).ReplaceAllString(string(data), "$1")), &v)
}

func HtmlToText(s string, limit int, unescape bool) string {
	var i int
	var tagStart bool
	var sigStart bool
	var sig []rune
	var runes = make([]rune, len(s))
	for _, r := range []rune(s) {
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
		case ';':
			if tagStart {
				break
			}
			if sigStart {
				if unescape {
					tr := []rune(html.UnescapeString(string(append([]rune{'&'}, append(sig, ';')...))))
					if len(tr) == 1 {
						runes[i] = tr[0]
						i++
					}
				} else {
					for _, r := range "&" + string(sig) + ";" {
						runes[i] = r
						i++
						limit++
					}
				}
			}
			sig = nil
			sigStart = false
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
			runes[i] = r
			i++
		}
		if i >= limit {
			break
		}
	}
	return string(runes[:i])
}

func SplitToLines(s string) (lines []string) {
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
	_, ext = SplitByLastByte(filename, '.')
	return
}

// PathClean has the same function with path.Clean(strings.ToLower(strings.Replace(strings.TrimSpace(s), "\\", "/", -1))),
// but it's faster!
func PathClean(path string) string {
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
			if c >= 'A' && c <= 'Z' {
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
