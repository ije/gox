package term

// ANSI Escape Sequences: Colours
// @see https://tldp.org/HOWTO/Bash-Prompt-HOWTO/x329.html
const (
	reset   = "\033[0m"
	black   = "\033[30m"
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
	white   = "\033[37m"
)

// Black returns a black colored string for terminal output.
func Black(msg string) string {
	return black + msg + reset
}

// Red returns a red colored string for terminal output.
func Red(msg string) string {
	return red + msg + reset
}

// Green returns a green colored string for terminal output.
func Green(msg string) string {
	return green + msg + reset
}

// Yellow returns a yellow colored string for terminal output.
func Yellow(msg string) string {
	return yellow + msg + reset
}

// Blue returns a blue colored string for terminal output.
func Blue(msg string) string {
	return blue + msg + reset
}

// Magenta returns a magenta colored string for terminal output.
func Magenta(msg string) string {
	return magenta + msg + reset
}

// Cyan returns a cyan colored string for terminal output.
func Cyan(msg string) string {
	return cyan + msg + reset
}

// White returns a white colored string for terminal output.
func White(msg string) string {
	return white + msg + reset
}

// Default returns a default colored string for terminal output.
func Default(msg string) string {
	return "\033[39m" + msg + reset
}

// Dim returns a grey colored string for terminal output.
func Dim(msg string) string {
	return "\033[2m" + msg + "\033[22m"
}
