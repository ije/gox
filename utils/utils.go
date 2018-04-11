package utils

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	B int64 = 1 << (10 * iota)
	KB
	MB
	GB
	TB
	PB
)

func WaitExit(callback func(os.Signal) bool, extraSignals ...os.Signal) {
	c := make(chan os.Signal, 1)
	signals := append([]os.Signal{os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT}, extraSignals...)
	signal.Notify(c, signals...)
	sig := <-c
	if callback != nil {
		if callback(sig) {
			os.Exit(1)
		} else {
			WaitExit(callback)
		}
	} else {
		os.Exit(1)
	}
}

func CopyFile(src string, dst string) (n int64, err error) {
	if src == dst {
		return
	}

	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	n, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return
	}

	return
}

func CopyDir(src string, dst string) (err error) {
	src = CleanPath(src)
	dst = CleanPath(dst)
	if src == dst {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return errors.New("source is not a directory")
	}

	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return
	}
	if err == nil {
		return errors.New("destination already exists")
	}

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return
	}

	for _, entry := range entries {
		srcPath := path.Join(src, entry.Name())
		dstPath := path.Join(dst, entry.Name())

		if entry.IsDir() {
			err = CopyDir(srcPath, dstPath)
			if err != nil {
				return
			}
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}

			_, err = CopyFile(srcPath, dstPath)
			if err != nil {
				return
			}
		}
	}

	return
}

func HashString(input interface{}, hasher hash.Hash) string {
	if hasher == nil {
		hasher = md5.New()
	}
	switch v := input.(type) {
	case string:
		hasher.Write([]byte(v))
	case []byte:
		hasher.Write(v)
	case io.Reader:
		io.Copy(hasher, v)
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func ParseBytes(s string) (int64, error) {
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

// todo: parse format '1d6h30m20s'
func ParseDuration(s string) (time.Duration, error) {
	if sl := len(s); sl > 0 {
		t := time.Second
		endWithS := false
	BeginParse:
		switch sl--; s[sl] {
		case 's', 'S':
			if sl == len(s)-1 {
				endWithS = true
				goto BeginParse
			}
		case 'n', 'N':
			if endWithS {
				t = time.Nanosecond
			}
		case 'µ', 'u', 'U':
			if endWithS {
				t = time.Microsecond
			}
		case 'm', 'M':
			if endWithS {
				t = time.Millisecond
			} else {
				t = time.Minute
			}
		case 'h', 'H':
			t = time.Hour
		case 'd', 'D':
			t = 24 * time.Hour
		default:
			sl++
		}
		if sl == 0 {
			return 0, strconv.ErrSyntax
		}

		f, err := strconv.ParseFloat(s[:sl], 64)
		if err != nil {
			return 0, strconv.ErrSyntax
		}

		t = time.Duration(float64(t) * f)
		return t, nil
	}

	return 0, strconv.ErrSyntax
}

func ParseJSONFile(filename string, v interface{}) (err error) {
	var r io.Reader
	if strings.HasPrefix(filename, "http://") || strings.HasPrefix(filename, "https://") {
		var resp *http.Response
		resp, err = http.Get(filename)
		if err != nil {
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			ret, _ := ioutil.ReadAll(resp.Body)
			err = fmt.Errorf("http(%d): %s"+resp.Status, string(ret))
			return
		}

		r = resp.Body
	} else {
		var file *os.File
		file, err = os.Open(filename)
		if err != nil {
			return
		}
		defer file.Close()

		r = file
	}

	return json.NewDecoder(r).Decode(v)
}

func SaveJSONFile(filename string, v interface{}) (err error) {
	return SaveJSONFileWithIndent(filename, v, "", "")
}

func SaveJSONFileWithIndent(filename string, v interface{}, prefix string, indent string) (err error) {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	je := json.NewEncoder(f)

	if len(indent) > 0 || len(prefix) > 0 {
		je.SetIndent(prefix, indent)
	}

	return je.Encode(v)
}

func DecodeGob(data []byte, v interface{}) (err error) {
	return gob.NewDecoder(bytes.NewReader(data)).Decode(v)
}

func EncodeGob(v interface{}) (data []byte, err error) {
	var buf = bytes.NewBuffer(nil)
	err = gob.NewEncoder(buf).Encode(v)
	if err != nil {
		return
	}

	data = buf.Bytes()
	return
}

func MustEncodeGob(v interface{}) []byte {
	if v == nil {
		return nil
	}

	var buf = bytes.NewBuffer(nil)
	err := gob.NewEncoder(buf).Encode(v)
	if err != nil {
		panic("gob: " + err.Error())
	}

	return buf.Bytes()
}

func ParseGobFile(filename string, v interface{}) (err error) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}

	defer f.Close()
	return gob.NewDecoder(f).Decode(v)
}

func SaveGobFile(filename string, v interface{}) (err error) {
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

func ParseLines(s string, keepEmptyLine bool) (lines []string) {
	for i, j, l := 0, 0, len(s); i < l; i++ {
		switch s[i] {
		case '\r', '\n':
			if i > j {
				lines = append(lines, s[j:i])
			} else if i == j && keepEmptyLine {
				lines = append(lines, "")
			}
			j = i + 1
			if s[i] == '\r' && i+1 < l && s[i+1] == '\n' {
				if i == l-2 && keepEmptyLine {
					lines = append(lines, "")
				}
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
		var err error
		f, err = strconv.ParseFloat(i, 64)
		if err != nil {
			ok = false
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

// CleanPath is the URL version of path.Clean, it returns a canonical URL path
// for p, eliminating . and .. elements.
//
// The following rules are applied iteratively until no further processing can
// be done:
//	1. Replace multiple slashes with a single slash.
//	2. Eliminate each . path name element (the current directory).
//	3. Eliminate each inner .. path name element (the parent directory)
//	   along with the non-.. element that precedes it.
//	4. Eliminate .. elements that begin a rooted path:
//	   that is, replace "/.." by "/" at the beginning of a path.
//
// If the result of this process is an empty string, "/" is returned
func CleanPath(p string) string {
	// Turn empty string into "/"
	if p == "" {
		return "/"
	}

	n := len(p)
	var buf []byte

	// Invariants:
	//      reading from path; r is index of next byte to process.
	//      writing to buf; w is index of next byte to write.

	// path must start with '/'
	r := 1
	w := 1

	if p[0] != '/' {
		r = 0
		buf = make([]byte, n+1)
		buf[0] = '/'
	}

	trailing := n > 2 && p[n-1] == '/'

	// A bit more clunky without a 'lazybuf' like the path package, but the loop
	// gets completely inlined (bufApp). So in contrast to the path package this
	// loop has no expensive function calls (except 1x make)

	for r < n {
		switch {
		case p[r] == '/':
			// empty path element, trailing slash is added after the end
			r++

		case p[r] == '.' && r+1 == n:
			trailing = true
			r++

		case p[r] == '.' && p[r+1] == '/':
			// . element
			r++

		case p[r] == '.' && p[r+1] == '.' && (r+2 == n || p[r+2] == '/'):
			// .. element: remove to last /
			r += 2

			if w > 1 {
				// can backtrack
				w--

				if buf == nil {
					for w > 1 && p[w] != '/' {
						w--
					}
				} else {
					for w > 1 && buf[w] != '/' {
						w--
					}
				}
			}

		default:
			// real path element.
			// add slash if needed
			if w > 1 {
				bufApp(&buf, p, w, '/')
				w++
			}

			// copy element
			for r < n && p[r] != '/' {
				bufApp(&buf, p, w, p[r])
				w++
				r++
			}
		}
	}

	// re-append trailing slash
	if trailing && w > 1 {
		bufApp(&buf, p, w, '/')
		w++
	}

	if buf == nil {
		return p[:w]
	}
	return string(buf[:w])
}

func bufApp(buf *[]byte, s string, w int, c byte) {
	if *buf == nil {
		if s[w] == c {
			return
		}

		*buf = make([]byte, len(s))
		copy(*buf, s[:w])
	}
	(*buf)[w] = c
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

func GetLocalIPs() (ips []string, err error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}

	if len(ips) == 0 {
		err = errors.New("not found")
	}
	return
}
