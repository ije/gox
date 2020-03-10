package tunnel

const (
	FlagHello Flag = iota + 1
	FlagProxy
	FlagReady
	FlagError
)

type Flag uint8

func (f Flag) String() string {
	switch f {
	case FlagHello:
		return "HELLO"
	case FlagProxy:
		return "PROXY"
	case FlagReady:
		return "READY"
	case FlagError:
		return "Error"
	default:
		return ""
	}
}
