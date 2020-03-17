package gwt

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"strings"
)

func encodeData(v interface{}, encoding string) (data []byte, err error) {
	buf := bytes.NewBuffer(nil)

	if encoding == "gob" {
		err = gob.NewEncoder(buf).Encode(v)
	} else {
		err = json.NewEncoder(buf).Encode(v)
	}
	if err != nil {
		return
	}

	data = buf.Bytes()
	return
}

func decodeData(data []byte, encoding string, v interface{}) error {
	r := bytes.NewReader(data)
	if encoding == "gob" {
		return gob.NewDecoder(r).Decode(v)
	}
	return json.NewDecoder(r).Decode(v)
}

// Encode GWT specific base64url encoding with padding stripped
func encodeSegment(seg []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(seg), "=")
}

// Decode GWT specific base64url encoding with padding stripped
func decodeSegment(seg string) ([]byte, error) {
	if l := len(seg) % 4; l > 0 {
		seg += strings.Repeat("=", 4-l)
	}
	return base64.URLEncoding.DecodeString(seg)
}
