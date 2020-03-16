package gwt

import (
	"bytes"
	"crypto"
	"crypto/hmac"
	"encoding/base64"
	"encoding/gob"
	"strings"
)

type expiresError string

func (e expiresError) Error() string {
	return string(e)
}

// IsExpires returns a boolean indicating whether the error is known to
// report that a gwt token is expired.
func IsExpires(err error) bool {
	_, ok := err.(expiresError)
	return ok
}

func sign(data []byte, secret string) string {
	hasher := hmac.New(crypto.SHA256.New, []byte(secret))
	hasher.Write(data)
	return encodeSegment(hasher.Sum(nil))
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

func encodeGob(v interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := gob.NewEncoder(buf).Encode(v)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
