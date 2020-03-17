package gwt

import (
	"time"
)

var defaultGWT = &GWT{
	Secret:   "gwt-secret",
	Encoding: "json",
}

// Config sets the secret and issuer of the default GWT
func Config(secret string, encoding string) {
	if len(secret) > 0 {
		defaultGWT.Secret = secret
	}
	if encoding == "json" || encoding == "gob" {
		defaultGWT.Encoding = encoding
	}
}

// SignToken creates a token with expires
func SignToken(payload interface{}, expires time.Duration) (token string, err error) {
	return defaultGWT.SignToken(payload, expires)
}

// ParseToken parses a token
func ParseToken(tokenString string, v interface{}) (err error) {
	return defaultGWT.ParseToken(tokenString, v)
}
