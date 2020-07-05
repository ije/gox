package gwt

type expiredError struct {
	message string
}

func (e *expiredError) Error() string {
	return e.message
}

// IsExpired returns a boolean indicating whether the error is known to
// report that a gwt token is expired.
func IsExpired(err error) bool {
	_, ok := err.(*expiredError)
	return ok
}
