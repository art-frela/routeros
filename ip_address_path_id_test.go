package routeros

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/art-frela/routeros/pkg/mockserver"
	"github.com/art-frela/routeros/types"
	"github.com/stretchr/testify/assert"
)

// mockClientTimeout bounds the plain http.Client used for raw mockserver
// probes (kept small and reasonable next to the generous shared testTimeout).
const mockClientTimeout = 5 * time.Second

// capturedRequest records the exact wire shape of the requests a service
// call produced (method, path, query, body) — evidence beyond status codes.
type capturedRequest struct {
	Calls       int
	Method      string
	Path        string // decoded path as seen by the server
	EscapedPath string // wire form of the path (raw * or %2A variant)
	RawQuery    string
	DotIDQuery  string
	Body        string
}

// newCaptureServer starts an httptest.Server that records every incoming
// request and always replies with the given status and canned body (an empty
// body mirrors the 204 No Content that live RouterOS answers DELETE with).
func newCaptureServer(t *testing.T, replyStatus int, replyBody string) (*httptest.Server, *capturedRequest) {
	t.Helper()

	captured := &capturedRequest{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
		}

		captured.Calls++
		captured.Method = r.Method
		captured.Path = r.URL.Path
		captured.EscapedPath = r.URL.EscapedPath()
		captured.RawQuery = r.URL.RawQuery
		captured.DotIDQuery = r.URL.Query().Get(".id")
		captured.Body = string(body)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(replyStatus)
		_, _ = w.Write([]byte(replyBody))
	}))
	t.Cleanup(ts.Close)

	return ts, captured
}

// newClientForServer builds a *Client aimed at the given test server through
// the envconfig-driven constructor (password falls back to the "master"
// default, matching mockserver.New("root", "master")).
func newClientForServer(t *testing.T, ts *httptest.Server) *Client {
	t.Helper()

	t.Setenv("ROS_TEST_FIND_BASE_URL", ts.URL)
	t.Setenv("ROS_TEST_FIND_USER", "root")

	cfg, err := NewClientConfigFromEnv("ROS_TEST_FIND")
	if err != nil {
		t.Fatalf("NewClientConfigFromEnv: %v", err)
	}

	c, err := NewClient(*cfg)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	return c
}

// mockRequest performs a raw authenticated request against a mockserver
// instance so tests control the exact wire form (path-form id, query-form
// id, literal JSON bodies with empty-string booleans).
func mockRequest(t *testing.T, ts *httptest.Server, method, rawPath, body string) (int, string) {
	t.Helper()

	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}

	req, err := http.NewRequestWithContext(context.Background(), method, ts.URL+rawPath, reader)
	if err != nil {
		t.Fatalf("build request %s %s: %v", method, rawPath, err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("root", "master")

	client := &http.Client{Timeout: mockClientTimeout}

	res, err := client.Do(req)
	if err != nil {
		t.Fatalf("do request %s %s: %v", method, rawPath, err)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	payload, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("read response of %s %s: %v", method, rawPath, err)
	}

	return res.StatusCode, string(payload)
}

// TestIPService_UpdateRemoveSendIDInPath proves the Amendment 3 contract:
// PATCH and DELETE must carry the resource id as a trailing path segment.
// Live RouterOS 7.23.3 honors the query form for GET only and answers
// query-form PATCH/DELETE with 400 "missing or invalid resource identifier"
// (curl probe matrices in .omo/evidence/task-3-*.md and task-4-*.md).
func TestIPService_UpdateRemoveSendIDInPath(t *testing.T) {
	tests := []struct {
		name           string
		call           func(ctx context.Context, svc *IPService) error
		wantMethod     string
		replyStatus    int
		replyBody      string
		wantBodySubstr string
	}{
		{
			name: "UpdateAddress sends PATCH to /ip/address/<id>",
			call: func(ctx context.Context, svc *IPService) error {
				_, err := svc.UpdateAddress(ctx, "*5", types.IPAddressAdd{
					Address:   "10.0.0.5/24",
					Interface: "ether1",
				})

				return err
			},
			wantMethod:     http.MethodPatch,
			replyStatus:    http.StatusOK,
			replyBody:      `{}`,
			wantBodySubstr: "10.0.0.5/24",
		},
		{
			name: "RemoveAddress sends DELETE to /ip/address/<id>",
			call: func(ctx context.Context, svc *IPService) error {
				return svc.RemoveAddress(ctx, "*5")
			},
			wantMethod:  http.MethodDelete,
			replyStatus: http.StatusNoContent, // live RouterOS: 204 with empty body
			replyBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// GIVEN a capture server recording the exact wire request
			ts, captured := newCaptureServer(t, tt.replyStatus, tt.replyBody)

			// WHEN the service mutates the *5 resource
			ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
			defer cancel()

			svc := &IPService{c: newClientForServer(t, ts)}

			if err := tt.call(ctx, svc); err != nil {
				t.Fatalf("service call: %v", err)
			}

			// THEN exactly one request went out, with the expected method
			assert.Equal(t, 1, captured.Calls, "expected exactly one API call")
			assert.Equal(t, tt.wantMethod, captured.Method)

			// ... the id as a trailing PATH segment (the decoded path is
			// independent of whether Go renders * raw or as %2A on the wire)
			assert.Equal(t, "/rest/ip/address/*5", captured.Path)
			assert.Contains(t,
				[]string{"/rest/ip/address/*5", "/rest/ip/address/%2A5"},
				captured.EscapedPath,
				"wire path must carry the id either raw or percent-escaped",
			)

			// ... and NO .id query parameter
			assert.Empty(t, captured.DotIDQuery)
			assert.Empty(t, captured.RawQuery)

			// ... and the request body contract (payload for PATCH, none for DELETE)
			if tt.wantBodySubstr != "" {
				assert.Contains(t, captured.Body, tt.wantBodySubstr)
			} else {
				assert.Empty(t, captured.Body)
			}
		})
	}
}

// TestIPService_GetAddressByIDKeepsQueryForm pins the GET-by-id contract:
// the query form stays (it is the one id form live RouterOS honors for GET)
// and the id must NOT migrate into the path.
func TestIPService_GetAddressByIDKeepsQueryForm(t *testing.T) {
	// GIVEN a capture server replying with a one-element address list
	ts, captured := newCaptureServer(t, http.StatusOK,
		`[{".id":"*5","address":"10.0.0.5/24","interface":"ether1"}]`)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	// WHEN the address *5 is fetched by id
	svc := &IPService{c: newClientForServer(t, ts)}
	got, err := svc.GetAddressByID(ctx, "*5")

	// THEN the request is a GET to the collection path with the .id QUERY
	assert.NoError(t, err)
	assert.Equal(t, http.MethodGet, captured.Method)
	assert.Equal(t, "/rest/ip/address", captured.Path)
	assert.Equal(t, "*5", captured.DotIDQuery)
	assert.Equal(t, "10.0.0.5/24", got.Address)
}

// TestMockserverIPAddressPathFormID proves mockserver parity with live
// RouterOS 7.23.3 for /ip/address: PATCH/DELETE honor the id as a trailing
// path segment, the legacy query form stays accepted, and bogus path ids
// are rejected with 4xx.
func TestMockserverIPAddressPathFormID(t *testing.T) {
	newMock := func(t *testing.T) *httptest.Server {
		t.Helper()

		dummy := mockserver.New("root", "master")
		mockserver.WithIPAddresses([]types.IPAddress{
			{
				ID:              "*1",
				Address:         "192.168.1.1/24",
				Network:         "192.168.1.0",
				Interface:       "ether1",
				ActualInterface: "ether1",
				Dynamic:         "false",
				Disabled:        "false",
			},
		})(dummy)

		ts := httptest.NewServer(http.HandlerFunc(dummy.IPAddresses))
		t.Cleanup(ts.Close)

		return ts
	}

	tests := []struct {
		name string
		run  func(t *testing.T, ts *httptest.Server)
	}{
		{
			name: "PATCH with path id updates the address",
			run: func(t *testing.T, ts *httptest.Server) {
				code, _ := mockRequest(t, ts, http.MethodPatch, "/rest/ip/address/*1",
					`{"address":"10.0.0.9/8","interface":"ether9"}`)
				assert.Equal(t, http.StatusOK, code)

				code, body := mockRequest(t, ts, http.MethodGet, "/rest/ip/address?.id=%2A1", "")
				assert.Equal(t, http.StatusOK, code)
				assert.Contains(t, body, "10.0.0.9/8")
			},
		},
		{
			name: "PATCH with percent-escaped path id behaves identically",
			run: func(t *testing.T, ts *httptest.Server) {
				code, _ := mockRequest(t, ts, http.MethodPatch, "/rest/ip/address/%2A1",
					`{"address":"10.0.0.10/8","interface":"ether9"}`)
				assert.Equal(t, http.StatusOK, code)

				_, body := mockRequest(t, ts, http.MethodGet, "/rest/ip/address?.id=%2A1", "")
				assert.Contains(t, body, "10.0.0.10/8")
			},
		},
		{
			name: "DELETE with path id removes the address",
			run: func(t *testing.T, ts *httptest.Server) {
				code, _ := mockRequest(t, ts, http.MethodDelete, "/rest/ip/address/*1", "")
				assert.Equal(t, http.StatusOK, code)

				code, body := mockRequest(t, ts, http.MethodGet, "/rest/ip/address?.id=%2A1", "")
				assert.Equal(t, http.StatusOK, code)
				assert.JSONEq(t, "[]", body)
			},
		},
		{
			name: "PATCH with bogus path id is rejected with 404",
			run: func(t *testing.T, ts *httptest.Server) {
				code, _ := mockRequest(t, ts, http.MethodPatch, "/rest/ip/address/*DEFACE", `{"comment":"x"}`)
				assert.Equal(t, http.StatusNotFound, code)
			},
		},
		{
			name: "DELETE with bogus path id is rejected with 404",
			run: func(t *testing.T, ts *httptest.Server) {
				code, _ := mockRequest(t, ts, http.MethodDelete, "/rest/ip/address/*DEFACE", "")
				assert.Equal(t, http.StatusNotFound, code)
			},
		},
		{
			name: "PATCH with query id stays accepted for backward compatibility",
			run: func(t *testing.T, ts *httptest.Server) {
				code, _ := mockRequest(t, ts, http.MethodPatch, "/rest/ip/address?.id=%2A1",
					`{"address":"10.0.0.11/8","interface":"ether9"}`)
				assert.Equal(t, http.StatusOK, code)

				_, body := mockRequest(t, ts, http.MethodGet, "/rest/ip/address?.id=%2A1", "")
				assert.Contains(t, body, "10.0.0.11/8")
			},
		},
		{
			name: "GET with path id returns the single address",
			run: func(t *testing.T, ts *httptest.Server) {
				code, body := mockRequest(t, ts, http.MethodGet, "/rest/ip/address/*1", "")
				assert.Equal(t, http.StatusOK, code)
				assert.Contains(t, body, `".id":"*1"`)
				assert.Contains(t, body, "192.168.1.1/24")
			},
		},
		{
			name: "PUT with path id is rejected (PUT targets the collection)",
			run: func(t *testing.T, ts *httptest.Server) {
				code, _ := mockRequest(t, ts, http.MethodPut, "/rest/ip/address/*1", `{"address":"10.0.0.12/8"}`)
				assert.Equal(t, http.StatusBadRequest, code)
			},
		},
		{
			name: "malformed path with extra segments is rejected",
			run: func(t *testing.T, ts *httptest.Server) {
				code, _ := mockRequest(t, ts, http.MethodGet, "/rest/ip/address/*1/extra", "")
				assert.Equal(t, http.StatusBadRequest, code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// GIVEN a fresh mock seeded with address *1
			// WHEN the raw request hits the handler
			// THEN the behavior mirrors live RouterOS (per-row asserts)
			tt.run(t, newMock(t))
		})
	}
}
