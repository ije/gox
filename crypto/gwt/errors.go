package gwt

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
