package utils

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

var (
	exitWaiting   bool
	exitCallbacks []func()
)

func CatchExit(callback func()) {
	if callback == nil {
		return
	}
	exitCallbacks = append(exitCallbacks, callback)

	if exitWaiting {
		return
	}
	exitWaiting = true

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

		for {
			fmt.Println("exit signal catched:", <-c)
			for _, callback := range exitCallbacks {
				callback()
			}
			os.Exit(1)
		}
	}()
}

func Contains(items interface{}, item interface{}) (ok bool) {
	switch a := items.(type) {
	case string:
		if len(a) == 0 {
			return
		}
		sep, yes := item.(string)
		ok = yes && strings.Index(a, sep) > -1
		return
	case []string:
		if len(a) == 0 {
			return
		}
		s, yes := item.(string)
		if yes {
			for _, c := range a {
				if c == s {
					ok = true
					return
				}
			}
		}
		return
	case []int:
		if len(a) == 0 {
			return
		}
		i, yes := item.(int)
		if yes {
			for _, c := range a {
				if c == i {
					return true
				}
			}
		}
		return
	default:
		return
	}
}

func CopyFile(src, dst string) (n int64, err error) {
	if src == dst {
		return
	}

	_, err = os.Lstat(dst)
	if err != nil && !os.IsNotExist(err) {
		return
	}

	sf, err := os.Open(src)
	if err != nil {
		return
	}
	defer sf.Close()

	df, err := os.Create(dst)
	if err != nil {
		return
	}
	defer df.Close()

	return io.Copy(df, sf)
}

func HashString(hasher, input interface{}) string {
	h, ok := hasher.(hash.Hash)
	if !ok {
		h = md5.New()
	}
	switch v := input.(type) {
	case []byte:
		h.Write(v)
	case string:
		h.Write([]byte(v))
	case io.Reader:
		io.Copy(h, v)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func UnmarshalJSONFile(filename string, v interface{}) (err error) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(v)
}

func MarshalJSONFile(filename string, v interface{}) (err error) {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(v)
}

func GetHttpJSON(url string, v interface{}) (err error) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = errors.New(resp.Status)
		return
	}

	err = json.NewDecoder(resp.Body).Decode(&v)
	return
}

func UnmarshalGobFile(filename string, v interface{}) (err error) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer f.Close()
	return gob.NewDecoder(f).Decode(v)
}

func MarshalGobFile(filename string, v interface{}) (err error) {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	return gob.NewEncoder(f).Encode(v)
}

func SplitByFirstByte(s string, c byte) (string, string) {
	for i, l := 0, len(s); i < l; i++ {
		if s[i] == c {
			return s[:i], s[i+1:]
		}
	}
	return s, ""
}

func SplitByLastByte(s string, c byte) (string, string) {
	for i := len(s) - 1; i >= 0; i-- {
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

func ToLines(s string) (lines []string) {
	for i, j, l := 0, 0, len(s); i < l; i++ {
		switch s[i] {
		case '\r', '\n':
			if i == j {
				lines = append(lines, "")
			} else if i > j {
				lines = append(lines, s[j:i])
			}
			j = i + 1
			if s[i] == '\r' && i+1 < l && s[i+1] == '\n' {
				i++
				j++
			}
		default:
			if i == l-1 && j < l {
				lines = append(lines, s[j:])
			}
		}
	}
	return
}

func ToNumber(v interface{}) (f float64, ok bool) {
	ok = true
	switch i := v.(type) {
	case string:
		i64, err := strconv.ParseInt(i, 10, 64)
		if err != nil {
			ok = false
		} else {
			f = float64(i64)
		}
	case int:
		f = float64(i)
	case int8:
		f = float64(i)
	case int16:
		f = float64(i)
	case int32:
		f = float64(i)
	case int64:
		f = float64(i)
	case uint:
		f = float64(i)
	case uint8:
		f = float64(i)
	case uint16:
		f = float64(i)
	case uint32:
		f = float64(i)
	case uint64:
		f = float64(i)
	case float32:
		f = float64(i)
	case float64:
		f = i
	default:
		ok = false
	}
	return
}

// PathClean has the same function with path.Clean(strings.Replace(strings.TrimSpace(s), "\\", "/", -1)),
// but it's faster!
func CleanPath(path string, toLower bool) string {
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

func GetLocalIp() (ip string, err error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip = ipnet.IP.String()
				break
			}
		}
	}

	if len(ip) == 0 {
		err = errors.New("not found")
	}
	return
}
