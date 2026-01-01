package term

func Bold(msg string) string {
	return "\033[1m" + msg + "\033[22m"
}

func Underline(msg string) string {
	return "\033[4m" + msg + "\033[24m"
}

func Italic(msg string) string {
	return "\033[3m" + msg + "\033[23m"
}

func Strikethrough(msg string) string {
	return "\033[9m" + msg + "\033[29m"
}
