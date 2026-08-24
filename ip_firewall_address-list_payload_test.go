package routeros

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/art-frela/routeros/pkg/mockserver"
	"github.com/art-frela/routeros/types"
	"github.com/stretchr/testify/assert"
)

// TestFirewallAddressListNewItemOmitEmptyOptionalFields proves the omitempty
// contract: a bare {Address, List} must marshal WITHOUT comment/disabled/
// dynamic keys. Live RouterOS 7.23.3 rejects the empty-string booleans the
// no-omitempty struct used to produce with 400 "invalid value of disabled,
// must be either yes or no" (probe matrix in .omo/evidence/task-5-*.md).
func TestFirewallAddressListNewItemOmitEmptyOptionalFields(t *testing.T) {
	tests := []struct {
		name     string
		item     types.FirewallAddressListNewItem
		wantJSON string
	}{
		{
			name: "bare item carries only address and list",
			item: types.FirewallAddressListNewItem{
				Address: "192.168.1.100",
				List:    "my-list",
			},
			wantJSON: `{"address":"192.168.1.100","list":"my-list"}`,
		},
		{
			name: "empty optional values are omitted too",
			item: types.FirewallAddressListNewItem{
				Address:  "192.168.1.100",
				Comment:  "",
				Disabled: "",
				Dynamic:  "",
				List:     "my-list",
			},
			wantJSON: `{"address":"192.168.1.100","list":"my-list"}`,
		},
		{
			name: "explicit optional values are still marshaled",
			item: types.FirewallAddressListNewItem{
				Address:  "192.168.1.100",
				Comment:  "blocked",
				Disabled: "no",
				Dynamic:  "no",
				List:     "my-list",
			},
			wantJSON: `{"address":"192.168.1.100","comment":"blocked","disabled":"no","dynamic":"no","list":"my-list"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// WHEN the new-item struct is marshaled
			raw, err := json.Marshal(tt.item)

			// THEN exactly the expected keys are present (and the optional
			// ones absent when unset)
			assert.NoError(t, err)
			assert.JSONEq(t, tt.wantJSON, string(raw))
		})
	}
}

// TestMockserverFirewallAddValidatesOptionalBooleans pins the mock to the
// live RouterOS probe matrix (task-5 evidence): absent disabled/dynamic
// keys are fine, yes/no/true/false are fine, anything else — including the
// empty string — is rejected with 400.
func TestMockserverFirewallAddValidatesOptionalBooleans(t *testing.T) {
	// GIVEN the mock firewall address-list handler
	dummy := mockserver.New("root", "master")
	ts := httptest.NewServer(http.HandlerFunc(dummy.IPFirewallAddressList))
	t.Cleanup(ts.Close)

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantDetail string
	}{
		{
			name:       "P4 bare address+list accepted",
			body:       `{"address":"192.168.1.100","list":"probe"}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "P3 disabled=no accepted",
			body:       `{"address":"192.168.1.101","list":"probe","disabled":"no"}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "true and false accepted",
			body:       `{"address":"192.168.1.102","list":"probe","disabled":"true","dynamic":"false"}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "yes accepted",
			body:       `{"address":"192.168.1.103","list":"probe","dynamic":"yes"}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "P1 empty disabled rejected like live RouterOS",
			body:       `{"address":"192.168.1.104","list":"probe","disabled":""}`,
			wantStatus: http.StatusBadRequest,
			wantDetail: "invalid value of disabled, must be either yes or no",
		},
		{
			name:       "P2 empty dynamic rejected like live RouterOS",
			body:       `{"address":"192.168.1.105","list":"probe","disabled":"no","dynamic":""}`,
			wantStatus: http.StatusBadRequest,
			wantDetail: "invalid value of dynamic, must be either yes or no",
		},
		{
			name:       "non-boolean value rejected",
			body:       `{"address":"192.168.1.106","list":"probe","disabled":"maybe"}`,
			wantStatus: http.StatusBadRequest,
			wantDetail: "invalid value of disabled, must be either yes or no",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// WHEN the raw payload hits the mock add handler
			code, body := mockRequest(t, ts, http.MethodPut, "/rest/ip/firewall/address-list", tt.body)

			// THEN the verdict mirrors live RouterOS
			assert.Equal(t, tt.wantStatus, code)
			if tt.wantDetail != "" {
				assert.Contains(t, body, tt.wantDetail)
			}
		})
	}
}

// TestIPFirewallAddressListServiceAddBareItem proves the omitempty fix end
// to end: a bare {Address, List} Add through the real client now passes the
// validating mock — no empty-string booleans on the wire anymore, so the
// Disabled/Dynamic workaround the integration suites carry is no longer
// required of callers.
func TestIPFirewallAddressListServiceAddBareItem(t *testing.T) {
	// GIVEN the validating mockserver behind the real client
	dummy := mockserver.New("root", "master")
	ts := httptest.NewServer(http.HandlerFunc(dummy.IPFirewallAddressList))
	t.Cleanup(ts.Close)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	svc := &IPFirewallAddressListService{c: newClientForServer(t, ts)}

	// WHEN a bare new item is added without any optional fields
	got, err := svc.Add(ctx, types.FirewallAddressListNewItem{
		Address: "zorro.com",
		List:    "bare",
	})

	// THEN the add succeeds and echoes the requested fields
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.NotEmpty(t, got.ID)
	assert.Equal(t, "zorro.com", got.Address)
	assert.Equal(t, "bare", got.List)
}
