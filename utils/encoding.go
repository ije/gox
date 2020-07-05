package utils

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

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
