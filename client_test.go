package routeros

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/art-frela/routeros/pkg/mockserver"
	"github.com/art-frela/routeros/types"
	"github.com/stretchr/testify/assert"
)

// recordingRoundTripper counts the requests passing through it
// and delegates the actual work to the default transport.
type recordingRoundTripper struct {
	hits int
}

func (r *recordingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	r.hits++

	return http.DefaultTransport.RoundTrip(req)
}

func TestNewClientHTTPClientInjection(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	// GIVEN a mock server with pre-configured IP addresses
	dummy := mockserver.New("root", "master")
	mockserver.WithIPAddresses([]types.IPAddress{
		{ID: "*1", Address: "192.168.1.1/24", Interface: "ether1", Disabled: "false"},
	})(dummy)

	ts := httptest.NewServer(http.HandlerFunc(dummy.IPAddresses))
	defer ts.Close()

	// GIVEN a client built with a custom *http.Client wrapping a recording round tripper
	// (a zero RequestTimeout expires the request context instantly, so set it explicitly)
	rec := &recordingRoundTripper{}
	custom := &http.Client{Transport: rec}

	c, err := NewClient(Config{
		BaseURL:        ts.URL,
		RequestTimeout: 5 * time.Second,
		User:           "root",
		Password:       "master",
		HTTPClient:     custom,
	})
	assert.NoError(t, err, "NewClient error")

	// WHEN the IPService is called to get addresses
	got, err := (&IPService{c: c}).GetAddresses(ctx)

	// THEN the injected HTTP client must serve the request
	assert.NoError(t, err, "GetAddresses error")
	assert.NotEmpty(t, got, "GetAddresses result")
	assert.Greater(t, rec.hits, 0, "injected *http.Client must be used for requests")
}

func TestNewClientNilHTTPClient(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	// GIVEN a mock server with pre-configured IP addresses
	dummy := mockserver.New("root", "master")
	mockserver.WithIPAddresses([]types.IPAddress{
		{ID: "*1", Address: "10.0.0.1/24", Interface: "ether2", Disabled: "false"},
	})(dummy)

	ts := httptest.NewServer(http.HandlerFunc(dummy.IPAddresses))
	defer ts.Close()

	// GIVEN a client built without an explicit HTTP client
	c, err := NewClient(Config{
		BaseURL:        ts.URL,
		RequestTimeout: 5 * time.Second,
		User:           "root",
		Password:       "master",
	})
	assert.NoError(t, err, "NewClient error")

	// THEN NewClient must fall back to its own non-nil HTTP client
	assert.NotNil(t, c.httpClient, "fallback *http.Client must be set")

	// WHEN the IPService is called to get addresses
	got, err := (&IPService{c: c}).GetAddresses(ctx)

	// THEN the request must succeed without a panic
	assert.NoError(t, err, "GetAddresses error")
	assert.Len(t, got, 1, "GetAddresses result")
}
