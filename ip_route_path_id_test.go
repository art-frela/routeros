package routeros

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/art-frela/routeros/pkg/mockserver"
	"github.com/art-frela/routeros/types"
	"github.com/stretchr/testify/assert"
)

// TestIPRouteService_UpdateRemoveSendIDInPath proves the Amendment 3
// contract for routes: PATCH and DELETE must carry the resource id as a
// trailing path segment. Live RouterOS 7.23.3 answers query-form
// PATCH/DELETE with 400 "missing or invalid resource identifier" (curl
// probe matrix in .omo/evidence/task-4-*.md).
func TestIPRouteService_UpdateRemoveSendIDInPath(t *testing.T) {
	tests := []struct {
		name          string
		call          func(ctx context.Context, svc *IPRouteService) error
		wantMethod    string
		replyStatus   int
		replyBody     string
		wantReqBody   string
		wantBodyEmpty bool
	}{
		{
			name: "UpdateRoute sends PATCH to /ip/route/<id>",
			call: func(ctx context.Context, svc *IPRouteService) error {
				_, err := svc.UpdateRoute(ctx, "*7", types.IPRouteAdd{
					DstAddress: "10.77.0.0/16",
					Gateway:    "192.168.77.1",
					Comment:    "it-updated",
				})

				return err
			},
			wantMethod:  http.MethodPatch,
			replyStatus: http.StatusOK,
			replyBody:   `{}`,
			wantReqBody: `{"dst-address":"10.77.0.0/16","gateway":"192.168.77.1","comment":"it-updated"}`,
		},
		{
			name: "RemoveRoute sends DELETE to /ip/route/<id>",
			call: func(ctx context.Context, svc *IPRouteService) error {
				return svc.RemoveRoute(ctx, "*7")
			},
			wantMethod:    http.MethodDelete,
			replyStatus:   http.StatusNoContent, // live RouterOS: 204 with empty body
			replyBody:     "",
			wantBodyEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// GIVEN a capture server recording the exact wire request
			ts, captured := newCaptureServer(t, tt.replyStatus, tt.replyBody)

			// WHEN the service mutates the *7 resource
			ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
			defer cancel()

			svc := &IPRouteService{c: newClientForServer(t, ts)}

			if err := tt.call(ctx, svc); err != nil {
				t.Fatalf("service call: %v", err)
			}

			// THEN exactly one request went out, with the expected method
			assert.Equal(t, 1, captured.Calls, "expected exactly one API call")
			assert.Equal(t, tt.wantMethod, captured.Method)

			// ... the id as a trailing PATH segment (decoded path is stable
			// whether Go renders * raw or as %2A on the wire)
			assert.Equal(t, "/rest/ip/route/*7", captured.Path)
			assert.Contains(t,
				[]string{"/rest/ip/route/*7", "/rest/ip/route/%2A7"},
				captured.EscapedPath,
				"wire path must carry the id either raw or percent-escaped",
			)

			// ... and NO .id query parameter
			assert.Empty(t, captured.DotIDQuery)
			assert.Empty(t, captured.RawQuery)

			// ... and the exact request body contract
			if tt.wantBodyEmpty {
				assert.Empty(t, captured.Body)
			} else {
				assert.JSONEq(t, tt.wantReqBody, captured.Body)
			}
		})
	}
}

// TestIPRouteService_GetRouteByIDKeepsQueryForm pins the GET-by-id
// contract: the query form stays (live RouterOS honors it for GET) and the
// id must NOT migrate into the path.
func TestIPRouteService_GetRouteByIDKeepsQueryForm(t *testing.T) {
	// GIVEN a capture server replying with a one-element route list
	ts, captured := newCaptureServer(t, http.StatusOK,
		`[{".id":"*7","dst-address":"10.77.0.0/16","gateway":"192.168.77.1"}]`)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	// WHEN the route *7 is fetched by id
	svc := &IPRouteService{c: newClientForServer(t, ts)}
	got, err := svc.GetRouteByID(ctx, "*7")

	// THEN the request is a GET to the collection path with the .id QUERY
	assert.NoError(t, err)
	assert.Equal(t, http.MethodGet, captured.Method)
	assert.Equal(t, "/rest/ip/route", captured.Path)
	assert.Equal(t, "*7", captured.DotIDQuery)
	assert.Equal(t, "10.77.0.0/16", got.DstAddress)
}

// TestMockserverIPRoutePathFormID proves mockserver parity with live
// RouterOS 7.23.3 for /ip/route: PATCH/DELETE honor the id as a trailing
// path segment, the legacy query form stays accepted, and bogus path ids
// are rejected with 4xx.
func TestMockserverIPRoutePathFormID(t *testing.T) {
	newMock := func(t *testing.T) *httptest.Server {
		t.Helper()

		dummy := mockserver.New("root", "master")
		mockserver.WithIPRoutes([]types.IPRoute{
			{
				ID:          "*1",
				DstAddress:  "10.0.0.0/8",
				Gateway:     "192.168.1.1",
				Distance:    "1",
				Scope:       "30",
				TargetScope: "10",
				Active:      "true",
				Static:      "true",
				Dynamic:     "false",
				Disabled:    "false",
			},
		})(dummy)

		ts := httptest.NewServer(http.HandlerFunc(dummy.IPRoutes))
		t.Cleanup(ts.Close)

		return ts
	}

	tests := []struct {
		name string
		run  func(t *testing.T, ts *httptest.Server)
	}{
		{
			name: "PATCH with path id updates the route",
			run: func(t *testing.T, ts *httptest.Server) {
				code, _ := mockRequest(t, ts, http.MethodPatch, "/rest/ip/route/*1",
					`{"dst-address":"10.0.0.0/8","gateway":"192.168.1.2","comment":"probe-p"}`)
				assert.Equal(t, http.StatusOK, code)

				code, body := mockRequest(t, ts, http.MethodGet, "/rest/ip/route?.id=%2A1", "")
				assert.Equal(t, http.StatusOK, code)
				assert.Contains(t, body, "probe-p")
			},
		},
		{
			name: "DELETE with path id removes the route",
			run: func(t *testing.T, ts *httptest.Server) {
				code, _ := mockRequest(t, ts, http.MethodDelete, "/rest/ip/route/*1", "")
				assert.Equal(t, http.StatusOK, code)

				code, body := mockRequest(t, ts, http.MethodGet, "/rest/ip/route?.id=%2A1", "")
				assert.Equal(t, http.StatusOK, code)
				assert.JSONEq(t, "[]", body)
			},
		},
		{
			name: "PATCH with bogus path id is rejected with 404",
			run: func(t *testing.T, ts *httptest.Server) {
				code, _ := mockRequest(t, ts, http.MethodPatch, "/rest/ip/route/*DEFACE", `{"comment":"x"}`)
				assert.Equal(t, http.StatusNotFound, code)
			},
		},
		{
			name: "DELETE with bogus path id is rejected with 404",
			run: func(t *testing.T, ts *httptest.Server) {
				code, _ := mockRequest(t, ts, http.MethodDelete, "/rest/ip/route/*DEFACE", "")
				assert.Equal(t, http.StatusNotFound, code)
			},
		},
		{
			name: "PATCH with query id stays accepted for backward compatibility",
			run: func(t *testing.T, ts *httptest.Server) {
				code, _ := mockRequest(t, ts, http.MethodPatch, "/rest/ip/route?.id=%2A1",
					`{"dst-address":"10.0.0.0/8","gateway":"192.168.1.3","comment":"probe-q"}`)
				assert.Equal(t, http.StatusOK, code)

				_, body := mockRequest(t, ts, http.MethodGet, "/rest/ip/route?.id=%2A1", "")
				assert.Contains(t, body, "probe-q")
			},
		},
		{
			name: "GET with path id returns the single route",
			run: func(t *testing.T, ts *httptest.Server) {
				code, body := mockRequest(t, ts, http.MethodGet, "/rest/ip/route/*1", "")
				assert.Equal(t, http.StatusOK, code)
				assert.Contains(t, body, `".id":"*1"`)
				assert.Contains(t, body, "10.0.0.0/8")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// GIVEN a fresh mock seeded with route *1
			// WHEN the raw request hits the handler
			// THEN the behavior mirrors live RouterOS (per-row asserts)
			tt.run(t, newMock(t))
		})
	}
}
