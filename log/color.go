package log

var (
	reset  = "\033[0m"
	grey   = "\033[90m"
	green  = "\033[32m"
	cyan   = "\033[36m"
	yellow = "\033[33m"
	red    = "\033[31m"
)

// colorize returns a colored string for terminal output.
func colorize(msg string, color string) string {
	return color + msg + reset
}

// Grey returns a grey colored string for terminal output.
func Grey(msg string) string {
	return colorize(msg, grey)
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

// Red returns a red colored string for terminal output.
func Red(msg string) string {
	return colorize(msg, red)
}
