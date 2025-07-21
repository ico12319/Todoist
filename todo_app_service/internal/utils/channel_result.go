package utils

type ChannelResult[T any] struct {
	Result T
	Err    error
}
