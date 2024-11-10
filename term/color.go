package term

// ANSI Escape Sequences: Colours
// @see https://tldp.org/HOWTO/Bash-Prompt-HOWTO/x329.html
var (
	reset   = "\033[0m"
	dim     = "\033[30m"
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
)

// Dim returns a grey colored string for terminal output.
func Dim(msg string) string {
	return dim + msg + reset
}

// Green returns a green colored string for terminal output.
func Green(msg string) string {
	return green + msg + reset
}

// Cyan returns a cyan colored string for terminal output.
func Cyan(msg string) string {
	return cyan + msg + reset
}

// Blue returns a blue colored string for terminal output.
func Blue(msg string) string {
	return blue + msg + reset
}

// Yellow returns a yellow colored string for terminal output.
func Yellow(msg string) string {
	return yellow + msg + reset
}

// Magenta returns a magenta colored string for terminal output.
func Magenta(msg string) string {
	return magenta + msg + reset
}

// Red returns a red colored string for terminal output.
func Red(msg string) string {
	return red + msg + reset
}
