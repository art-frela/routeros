package routeros

import (
	"errors"
	"fmt"
)

var (
	// ErrMissBaseURL is returned by NewClient when Config.BaseURL is empty.
	ErrMissBaseURL = errors.New("miss baseURL")
	// ErrMissUserOrPass is returned by NewClient when Config.User
	// or Config.Password is empty.
	ErrMissUserOrPass = errors.New("miss user/password")
)

// ResponseError represents a non-successful HTTP response (status >= 400)
// from the RouterOS REST API. Body holds the raw response payload.
type ResponseError struct {
	StatusCode int
	Body       string
}

// Error implements the error interface, preserving the legacy
// "status_code: %d, response: %s" format byte-identical.
func (e *ResponseError) Error() string {
	return fmt.Sprintf("status_code: %d, response: %s", e.StatusCode, e.Body)
}
