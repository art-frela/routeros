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

func TestIPService_GetAddresses(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

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
		{
			ID:              "*2",
			Address:         "10.0.0.1/8",
			Network:         "10.0.0.0",
			Interface:       "ether2",
			ActualInterface: "ether2",
			Dynamic:         "false",
			Disabled:        "true",
		},
	})(dummy)

	ts := httptest.NewServer(http.HandlerFunc(dummy.IPAddresses))
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

	ips := &IPService{c: c}

	got, err := ips.GetAddresses(ctx)

	assert.NoError(t, err)
	assert.Len(t, got, 2)
	assert.Equal(t, "192.168.1.1/24", got[0].Address)
	assert.Equal(t, "10.0.0.1/8", got[1].Address)
}

func TestIPService_GetAddressByID(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

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
		{
			ID:              "*2",
			Address:         "10.0.0.1/8",
			Network:         "10.0.0.0",
			Interface:       "ether2",
			ActualInterface: "ether2",
			Dynamic:         "false",
			Disabled:        "true",
		},
	})(dummy)

	ts := httptest.NewServer(http.HandlerFunc(dummy.IPAddresses))
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

	ips := &IPService{c: c}

	got, err := ips.GetAddressByID(ctx, "*1")

	assert.NoError(t, err)
	assert.Equal(t, "192.168.1.1/24", got.Address)
	assert.Equal(t, "ether1", got.Interface)
}

func TestIPService_AddAddress(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

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

	ips := &IPService{c: c}

	newAddr := types.IPAddressAdd{
		Address:   "10.0.0.2/8",
		Interface: "ether2",
	}

	got, err := ips.AddAddress(ctx, newAddr)

	assert.NoError(t, err)
	assert.Equal(t, "10.0.0.2/8", got.Address)
	assert.Equal(t, "ether2", got.Interface)
}

func TestIPService_RemoveAddress(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

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

	ips := &IPService{c: c}

	err = ips.RemoveAddress(ctx, "*1")

	assert.NoError(t, err)
}

func TestIPService_UpdateAddress(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

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

	ips := &IPService{c: c}

	updatedAddr := types.IPAddressAdd{
		Address:   "192.168.1.2/24",
		Interface: "ether2",
		Disabled:  "true",
	}

	got, err := ips.UpdateAddress(ctx, "*1", updatedAddr)

	assert.NoError(t, err)
	assert.Equal(t, "192.168.1.2/24", got.Address)
	assert.Equal(t, "ether2", got.Interface)
	assert.Equal(t, "true", got.Disabled)
}
