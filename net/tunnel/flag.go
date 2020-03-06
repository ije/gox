package tunnel

const (
	FlagHello Flag = iota + 1
	FlagProxy
)

type Flag uint8
