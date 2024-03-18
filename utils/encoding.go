package utils

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"os"
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
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()
	return json.NewDecoder(file).Decode(v)
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
