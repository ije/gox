package gwt

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"

	"gox/utils"
)

// GWT is an api token rsa crypto with gob encoding
type GWT struct {
	Secret string
}

// Channel is a payload container includes expires and issuer info
type Channel struct {
	Payload   []byte
	ExpiresAt int64
	Issuer    string
}

// SignToken creates a token with expires
func (gwt *GWT) SignToken(payload interface{}, expires time.Duration) (token string, err error) {
	payloadData, err := encodeGob(payload)
	if err != nil {
		err = fmt.Errorf("can not encode payload: %v", err)
		return
	}

	chData, err := encodeGob(Channel{
		Payload:   payloadData,
		ExpiresAt: time.Now().Add(expires).Unix(),
		Issuer:    "GWT",
	})
	if err != nil {
		err = fmt.Errorf("can not encode channel: %v", err)
		return
	}

	return encodeSegment(chData) + "." + sign(chData, gwt.Secret), nil
}

// ParseToken parses a token
func (gwt *GWT) ParseToken(tokenString string, v interface{}) (err error) {
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
	err = gob.NewDecoder(bytes.NewReader(chData)).Decode(&ch)
	if err != nil {
		err = fmt.Errorf("bad channel data")
		return
	}

	d := time.Now().Unix() - ch.ExpiresAt
	if d > 0 {
		err = expiresError(fmt.Sprintf("token is expired by %v", time.Duration(d)*time.Second))
		return
	}

	if ch.Issuer != "GWT" {
		err = fmt.Errorf("invalid issuer '%s'", ch.Issuer)
		return
	}

	err = gob.NewDecoder(bytes.NewReader(ch.Payload)).Decode(v)
	if err != nil {
		err = fmt.Errorf("bad payload data")
	}
	return
}
