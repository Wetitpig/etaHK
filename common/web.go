package common

const MAX_CONN = 64

type JsonRetMsg[T any] struct {
	UID int
	Ret T
}
