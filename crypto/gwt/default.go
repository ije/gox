package gwt

import (
	"time"
)

var defaultGWT = &GWT{
	Secret: "gwt-secret",
}

// Config sets the secret and issuer of the default GWT
func Config(secret string) {
	defaultGWT.Secret = secret
}

// SignToken creates a token with expires
func SignToken(payload interface{}, expires time.Duration) (token string, err error) {
	return defaultGWT.SignToken(payload, expires)
}

// ParseToken parses a token
func ParseToken(tokenString string, v interface{}) (err error) {
	return defaultGWT.ParseToken(tokenString, v)
}
