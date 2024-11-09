package term

var (
	reset   = "\033[0m"
	dim     = "\033[30m"
	green   = "\033[32m"
	cyan    = "\033[36m"
	yellow  = "\033[33m"
	magenta = "\033[35m"
	red     = "\033[31m"
)

// colorize returns a colored string for terminal output.
func colorize(msg string, color string) string {
	return color + msg + reset
}

// Dim returns a grey colored string for terminal output.
func Dim(msg string) string {
	return colorize(msg, dim)
}

// Green returns a green colored string for terminal output.
func Green(msg string) string {
	return colorize(msg, green)
}

// Cyan returns a cyan colored string for terminal output.
func Cyan(msg string) string {
	return colorize(msg, cyan)
}

// Yellow returns a yellow colored string for terminal output.
func Yellow(msg string) string {
	return colorize(msg, yellow)
}

// Magenta returns a magenta colored string for terminal output.
func Magenta(msg string) string {
	return colorize(msg, magenta)
}

// Red returns a red colored string for terminal output.
func Red(msg string) string {
	return colorize(msg, red)
}
