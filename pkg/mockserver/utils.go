package mockserver

import (
	"math/rand"
)

const keyLen = 6

var (
	letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
)

func newKey() string {
	return "*" + randStringRunes(keyLen)
}

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	return string(b)
}

func Ptr[T any](v T) *T {
	return &v
}
