package routeros

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/art-frela/routeros/pkg/mockserver"
	"github.com/art-frela/routeros/types"
	"github.com/stretchr/testify/assert"
)

// errorsTestTimeout keeps error-path tests fast instead of the shared testTimeout.
const errorsTestTimeout = 5 * time.Second

func TestNewClient_SentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want error
	}{
		{
			name: "empty BaseURL matches ErrMissBaseURL",
			cfg:  Config{User: "root", Password: "master"},
			want: ErrMissBaseURL,
		},
		{
			name: "empty User matches ErrMissUserOrPass",
			cfg:  Config{BaseURL: "http://192.0.2.1", Password: "master"},
			want: ErrMissUserOrPass,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// GIVEN a config with a missing mandatory field
			// WHEN the client is constructed
			_, err := NewClient(tt.cfg)

			// THEN the error matches the exported sentinel via errors.Is
			assert.ErrorIs(t, err, tt.want, "NewClient sentinel error")
		})
	}
}

func TestMakeRequest_HTTPErrorAsResponseError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), errorsTestTimeout)
	defer cancel()

	// GIVEN a mockserver-backed test server and a client with wrong credentials
	dummy := mockserver.New("root", "master")
	ts := httptest.NewServer(http.HandlerFunc(dummy.IPAddresses))
	defer ts.Close()

	c, err := NewClient(Config{
		BaseURL:              ts.URL,
		RequestTimeout:       errorsTestTimeout,
		PauseBetweenRequests: time.Millisecond,
		BurstRequestCount:    1,
		User:                 "root",
		Password:             "wrong",
	})
	assert.NoError(t, err, "NewClient error")

	// WHEN a service method is called and the mock rejects basic auth with 401
	_, err = (&IPService{c: c}).GetAddresses(ctx)

	// THEN the error unwraps to *ResponseError carrying the 401 status code
	var re *ResponseError
	if assert.True(t, errors.As(err, &re), "error should unwrap to *ResponseError") {
		assert.Equal(t, http.StatusUnauthorized, re.StatusCode, "ResponseError.StatusCode")

		// AND the body is byte-identical to what the mock wrote (same encoder, same struct),
		// AND Error() preserves the legacy format byte-identical.
		var wantBody bytes.Buffer
		assert.NoError(t, json.NewEncoder(&wantBody).Encode(&types.Error{
			Error:   http.StatusUnauthorized,
			Message: http.StatusText(http.StatusUnauthorized),
		}), "encode expected mock body")

		assert.Equal(t, wantBody.String(), re.Body, "ResponseError.Body must be the raw mock response body")
		assert.Equal(
			t,
			fmt.Sprintf("status_code: %d, response: %s", http.StatusUnauthorized, wantBody.String()),
			re.Error(),
			"ResponseError.Error() must keep the legacy format byte-identical",
		)
	}
}

func TestMakeRequest_DecodeErrorWrapped(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), errorsTestTimeout)
	defer cancel()

	// GIVEN a plain test server (not mockserver) answering 200 OK with a non-JSON body
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "not-json{")
	}))
	defer ts.Close()

	c, err := NewClient(Config{
		BaseURL:              ts.URL,
		RequestTimeout:       errorsTestTimeout,
		PauseBetweenRequests: time.Millisecond,
		BurstRequestCount:    1,
		User:                 "root",
		Password:             "master",
	})
	assert.NoError(t, err, "NewClient error")

	// WHEN a service method gets the malformed body decoded
	_, err = (&IPService{c: c}).GetAddresses(ctx)

	// THEN the JSON decode error is reported, prefixed, and unwrappable via errors.Unwrap
	assert.Error(t, err, "GetAddresses should fail on a non-JSON body")
	assert.ErrorContains(t, err, "decode response", "wrapped error keeps the decode prefix")
	assert.NotNil(t, errors.Unwrap(err), "decode error must be wrapped with %w")
}
