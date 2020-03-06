package tunnel

const (
	FlagHello Flag = iota + 1
	FlagProxy
)

type Flag uint8

func (f Flag) String() string {
	switch f {
	case FlagHello:
		return "HELLO"
	case FlagProxy:
		return "PROXY"
	default:
		return ""
	}
}
