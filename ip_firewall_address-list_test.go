package routeros

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIPFirewallAddressListService_Find(t *testing.T) {
	cfg, err := NewClientConfigFromEnv("ROS_TEST")
	if err != nil {
		t.Errorf("NewClientConfigFromEnv error: %s", err)

		return
	}

	c, err := NewClient(*cfg)
	if err != nil {
		t.Errorf("NewClient error: %s", err)

		return
	}

	dt2ip, _ := time.Parse(time.DateTime, "2025-04-06 02:23:27")

	tests := []struct {
		name    string
		c       *Client
		list    string
		ips     []string
		want    FirewallAddressList
		wantErr bool
	}{
		{
			name: "test.1 handly run - 2ip.ru",
			c:    c,
			list: "my-address-list",
			ips:  []string{"2ip.ru"},
			want: FirewallAddressList{
				{
					ID:           "*2D9AC9",
					Address:      "2ip.ru",
					CreationTime: DateTime{Time: dt2ip},
					Disabled:     "false",
					Dynamic:      "false",
					List:         "my-address-list",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipfwls := &IPFirewallAddressListService{
				c: tt.c,
			}
			got, err := ipfwls.Find(t.Context(), tt.list, tt.ips...)

			if tt.wantErr {
				assert.Error(t, err, "IPFirewallAddressListService.Find() error")
			} else {
				assert.NoError(t, err, "IPFirewallAddressListService.Find() error")
			}

			assert.Equal(t, tt.want, got, "IPFirewallAddressListService.Find() list")
		})
	}
}

func TestIPFirewallAddressListService_Add(t *testing.T) {
	cfg, err := NewClientConfigFromEnv("ROS_TEST")
	if err != nil {
		t.Errorf("NewClientConfigFromEnv error: %s", err)

		return
	}

	c, err := NewClient(*cfg)
	if err != nil {
		t.Errorf("NewClient error: %s", err)

		return
	}

	tests := []struct {
		name    string
		c       *Client
		item    FirewallAddressListNewItem
		wantErr bool
	}{
		{
			name: "test.1 handly run",
			c:    c,
			item: FirewallAddressListNewItem{
				Address:  "zorro.com",
				Disabled: "false",
				Dynamic:  "false",
				List:     "my-address-list",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipfwls := &IPFirewallAddressListService{
				c: tt.c,
			}

			if err := ipfwls.Add(t.Context(), tt.item); (err != nil) != tt.wantErr {
				t.Errorf("IPFirewallAddressListService.Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
