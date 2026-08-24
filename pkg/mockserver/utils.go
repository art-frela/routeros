package mockserver

import (
	"errors"
	"math/rand"
	"net/http"

	"github.com/art-frela/routeros/types"
)

const keyLen = 6

// errResourceNotFound marks lookups of unknown resource ids so handlers can
// answer 404 like real RouterOS does for path-form ids.
var errResourceNotFound = errors.New("resource not found")

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

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

// writeResourceError writes a RouterOS-style error envelope, mapping
// errResourceNotFound to 404 and anything else to 500.
func writeResourceError(w http.ResponseWriter, err error) {
	code := http.StatusInternalServerError
	if errors.Is(err, errResourceNotFound) {
		code = http.StatusNotFound
	}

	writeResponseJSON(w, code, types.Error{
		Detail:  err.Error(),
		Error:   code,
		Message: http.StatusText(code),
	})
}
