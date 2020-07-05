// Package gwt implements an api token manager with rsa crypto
package gwt

import (
	"crypto"
	"crypto/hmac"
	"crypto/sha1"
	"fmt"
	"time"

	"github.com/ije/gox/utils"
)

// GWT is an api token manager with rsa crypto
type GWT struct {
	Secret   string
	Encoding string
}

// Channel is a payload container includes expires and issuer for GWT
type Channel struct {
	Payload   []byte
	ExpiresAt int64
	Issuer    string
}

// SignToken creates a token
func (gwt *GWT) SignToken(payload interface{}, expires time.Duration) (token string, err error) {
	return gwt.SignTokenBy("GWT", payload, expires)
}

// SignTokenBy creates a token with issuer
func (gwt *GWT) SignTokenBy(issuer string, payload interface{}, expires time.Duration) (token string, err error) {
	payloadData, err := encodeData(payload, gwt.Encoding)
	if err != nil {
		err = fmt.Errorf("can not encode payload: %v", err)
		return
	}

	chData, err := encodeData(Channel{
		Payload:   payloadData,
		ExpiresAt: time.Now().UTC().Add(expires).Unix(),
		Issuer:    issuer,
	}, gwt.Encoding)
	if err != nil {
		err = fmt.Errorf("can not encode channel: %v", err)
		return
	}

	return encodeSegment(chData) + "." + sign(chData, gwt.Secret), nil
}

// ParseToken parses a token
func (gwt *GWT) ParseToken(tokenString string, v interface{}) (err error) {
	return gwt.ParseTokenBy("GWT", tokenString, v)
}

// ParseTokenBy parses a token with issuer
func (gwt *GWT) ParseTokenBy(issuer string, tokenString string, v interface{}) (err error) {
	p1, sig := utils.SplitByFirstByte(tokenString, '.')
	chData, err := decodeSegment(p1)
	if err != nil {
		return
	}

	if sign(chData, gwt.Secret) != sig {
		err = fmt.Errorf("invalid signature")
		return
	}

	var ch Channel
	err = decodeData(chData, gwt.Encoding, &ch)
	if err != nil {
		err = fmt.Errorf("bad channel data")
		return
	}

	if ch.Issuer != issuer {
		err = fmt.Errorf("invalid issuer '%s'", ch.Issuer)
		return
	}

	d := time.Now().UTC().Unix() - ch.ExpiresAt
	if d > 0 {
		err = &expiredError{fmt.Sprintf("token is expired by %v", time.Duration(d)*time.Second)}
		return
	}

	err = decodeData(ch.Payload, gwt.Encoding, v)
	if err != nil {
		err = fmt.Errorf("bad payload data")
	}
	return
}

func sign(data []byte, secret string) string {
	hasher := hmac.New(crypto.SHA256.New, []byte(secret))
	sha := sha1.New()
	sha.Write(data)
	hasher.Write(sha.Sum(nil))
	return encodeSegment(hasher.Sum(nil))
}
