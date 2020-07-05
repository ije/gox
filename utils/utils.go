package utils

import (
	"archive/zip"
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

func WaitExitSignal(callback func(os.Signal) bool) {
	if callback == nil {
		return
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGHUP)
	if !callback(<-c) {
		WaitExitSignal(callback)
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

func ZipTo(path string, output io.Writer) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}

	if fi.IsDir() {
		absDir, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		archive := zip.NewWriter(output)
		defer archive.Close()

		return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}

			header.Name = strings.TrimPrefix(strings.TrimPrefix(path, absDir), "/")
			if header.Name == "" {
				return nil
			}

			if info.IsDir() {
				header.Name += "/"
			} else {
				header.Method = zip.Deflate
			}

			writer, err := archive.CreateHeader(header)
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(writer, file)
			return err
		})
	}

	header, err := zip.FileInfoHeader(fi)
	if err != nil {
		return err
	}
	header.Method = zip.Deflate

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	archive := zip.NewWriter(output)
	defer archive.Close()

	gzw, err := archive.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(gzw, file)
	return err
}

func MustEncodeJSON(v interface{}) []byte {
	buf := bytes.NewBuffer(nil)
	err := json.NewEncoder(buf).Encode(v)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
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

func WriteJSONFile(filename string, v interface{}, indent string) (err error) {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	je := json.NewEncoder(f)

	if len(indent) > 0 {
		je.SetIndent("", indent)
	}

	return je.Encode(v)
}

func MustEncodeGob(v interface{}) []byte {
	buf := bytes.NewBuffer(nil)
	err := gob.NewEncoder(buf).Encode(v)
	if err != nil {
		panic(err)
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

func WriteGobFile(filename string, v interface{}) (err error) {
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

func ToNumber(v interface{}) (f float64, err error) {
	switch i := v.(type) {
	case string:
		f, err = strconv.ParseFloat(i, 64)
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
		err = errors.New("NaN")
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

// GetLocalIPList return the list of local ip address
func GetLocalIPList() (list []string, err error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				list = append(list, ipnet.IP.String())
			}
		}
	}

	if len(list) == 0 {
		err = errors.New("not found")
	}
	return
}
