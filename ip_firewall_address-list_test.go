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

const testTimeout = time.Second * 150000

func TestIPFirewallAddressListService_Find(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	dt2ip, _ := time.Parse(time.DateTime, "2025-04-06 02:23:27")

	dummy := mockserver.New("root", "master")
	mockserver.WithIPFireWallAddressList(map[string]types.FirewallAddressList{
		"test": {
			{
				ID:           "*2D9AC9",
				Address:      "2ip.ru",
				CreationTime: types.DateTime{Time: dt2ip},
				Disabled:     "false",
				Dynamic:      "false",
				List:         "test",
			},
		},
	})(dummy)

	ts := httptest.NewServer(http.HandlerFunc(dummy.IPFirewallAddressList))
	defer ts.Close()

	t.Setenv("ROS_TEST_FIND_BASE_URL", ts.URL)
	t.Setenv("ROS_TEST_FIND_USER", "root")

	cfg, err := NewClientConfigFromEnv("ROS_TEST_FIND")
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
		list    string
		ips     []string
		want    types.FirewallAddressList
		wantErr bool
	}{
		{
			name: "test.1 ok, exists - 2ip.ru",
			c:    c,
			list: "test",
			ips:  []string{"2ip.ru"},
			want: types.FirewallAddressList{
				{
					ID:           "*2D9AC9",
					Address:      "2ip.ru",
					CreationTime: types.DateTime{Time: dt2ip},
					Disabled:     "false",
					Dynamic:      "false",
					List:         "test",
				},
			},
			wantErr: false,
		},
		{
			name:    "test.2 ok, not exists",
			c:       c,
			list:    "test",
			ips:     []string{"7ip.ru"},
			want:    types.FirewallAddressList{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipfwls := &IPFirewallAddressListService{
				c: tt.c,
			}
			got, err := ipfwls.Find(ctx, tt.list, tt.ips...)

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
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	dt2ip, _ := time.Parse(time.DateTime, "2025-04-06 02:23:27")

	dummy := mockserver.New("root", "master")
	mockserver.WithIPFireWallAddressList(map[string]types.FirewallAddressList{
		"test": {
			{
				ID:           "*2D9AC9",
				Address:      "2ip.ru",
				CreationTime: types.DateTime{Time: dt2ip},
				Disabled:     "false",
				Dynamic:      "false",
				List:         "test",
			},
		},
	})(dummy)

	ts := httptest.NewServer(http.HandlerFunc(dummy.IPFirewallAddressList))
	defer ts.Close()

	t.Setenv("ROS_TEST_FIND_BASE_URL", ts.URL)
	t.Setenv("ROS_TEST_FIND_USER", "root")

	cfg, err := NewClientConfigFromEnv("ROS_TEST_FIND")
	if err != nil {
		t.Errorf("NewClientConfigFromEnv error: %s", err)

		return
	}

	c, err := NewClient(*cfg)
	if err != nil {
		t.Errorf("NewClient error: %s", err)

		return
	}

	_, offset := time.Now().Zone()

	tests := []struct {
		name    string
		c       *Client
		item    types.FirewallAddressListNewItem
		want    *types.FirewallAddressListItem
		wantErr bool
	}{
		{
			name: "test.1 ok - add new",
			c:    c,
			item: types.FirewallAddressListNewItem{
				Address:  "zorro.com",
				Disabled: "false",
				Dynamic:  "false",
				List:     "test",
			},
			want: &types.FirewallAddressListItem{
				Address:      "zorro.com",
				CreationTime: types.DateTime{Time: time.Now().Add(time.Duration(offset) * time.Second).Truncate(time.Minute).In(time.UTC)},
				Disabled:     "false",
				Dynamic:      "false",
				List:         "test",
			},
			wantErr: false,
		},
		{
			name: "test.2 err - add exists",
			c:    c,
			item: types.FirewallAddressListNewItem{
				Address:  "zorro.com",
				Disabled: "false",
				Dynamic:  "false",
				List:     "test",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipfwls := &IPFirewallAddressListService{
				c: tt.c,
			}

			got, err := ipfwls.Add(ctx, tt.item)
			if got != nil {
				got.ID = tt.want.ID
			}

			if tt.wantErr {
				assert.Error(t, err, "IPFirewallAddressListService.Add() error")
			} else {
				assert.NoError(t, err, "IPFirewallAddressListService.Add() error")
			}

			assert.Equal(t, tt.want, got, "IPFirewallAddressListService.Add() new item")
		})
	}
}
