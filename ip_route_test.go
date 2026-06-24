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

func TestIPRouteService_GetRoutes(t *testing.T) {
	// GIVEN a mock server with pre-configured IP routes
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	dummy := mockserver.New("root", "master")
	mockserver.WithIPRoutes([]types.IPRoute{
		{
			ID:          "*1",
			DstAddress:  "0.0.0.0/0",
			Gateway:     "10.155.101.1",
			Distance:    "1",
			Scope:       "30",
			TargetScope: "10",
			Active:      "true",
			Static:      "true",
			Dynamic:     "false",
			Disabled:    "false",
		},
		{
			ID:          "*2",
			DstAddress:  "192.168.1.0/24",
			Gateway:     "ether2",
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

	// WHEN the IPRouteService is called to get routes
	svc := &IPRouteService{c: c}
	got, err := svc.GetRoutes(ctx)

	// THEN it should return the expected routes
	assert.NoError(t, err)
	assert.Len(t, got, 2)
	assert.Equal(t, "0.0.0.0/0", got[0].DstAddress)
	assert.Equal(t, "10.155.101.1", got[0].Gateway)
	assert.Equal(t, "192.168.1.0/24", got[1].DstAddress)
	assert.Equal(t, "ether2", got[1].Gateway)
}

func TestIPRouteService_GetRouteByID(t *testing.T) {
	// GIVEN a mock server with a specific IP route
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	dummy := mockserver.New("root", "master")
	mockserver.WithIPRoutes([]types.IPRoute{
		{
			ID:          "*1",
			DstAddress:  "0.0.0.0/0",
			Gateway:     "10.155.101.1",
			Distance:    "1",
			Scope:       "30",
			TargetScope: "10",
			Active:      "true",
			Static:      "true",
			Dynamic:     "false",
			Disabled:    "false",
		},
		{
			ID:          "*2",
			DstAddress:  "10.0.0.0/8",
			Gateway:     "ether3",
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

	// WHEN the IPRouteService is called to get a route by ID
	svc := &IPRouteService{c: c}
	got, err := svc.GetRouteByID(ctx, "*1")

	// THEN it should return the expected route
	assert.NoError(t, err)
	assert.Equal(t, "0.0.0.0/0", got.DstAddress)
	assert.Equal(t, "10.155.101.1", got.Gateway)
}

func TestIPRouteService_AddRoute(t *testing.T) {
	// GIVEN a mock server with an existing route
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	dummy := mockserver.New("root", "master")
	mockserver.WithIPRoutes([]types.IPRoute{
		{
			ID:          "*1",
			DstAddress:  "0.0.0.0/0",
			Gateway:     "10.155.101.1",
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

	// WHEN a new route is added
	svc := &IPRouteService{c: c}
	newRoute := types.IPRouteAdd{
		DstAddress: "10.0.0.0/8",
		Gateway:    "ether3",
		Distance:   "1",
	}

	got, err := svc.AddRoute(ctx, newRoute)

	// THEN it should return the created route with the expected fields
	assert.NoError(t, err)
	assert.Equal(t, "10.0.0.0/8", got.DstAddress)
	assert.Equal(t, "ether3", got.Gateway)
	assert.Equal(t, "true", got.Static)
}

func TestIPRouteService_RemoveRoute(t *testing.T) {
	// GIVEN a mock server with an existing route
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	dummy := mockserver.New("root", "master")
	mockserver.WithIPRoutes([]types.IPRoute{
		{
			ID:          "*1",
			DstAddress:  "0.0.0.0/0",
			Gateway:     "10.155.101.1",
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

	// WHEN the route is removed by ID
	svc := &IPRouteService{c: c}
	err = svc.RemoveRoute(ctx, "*1")

	// THEN it should succeed without error
	assert.NoError(t, err)
}

func TestIPRouteService_UpdateRoute(t *testing.T) {
	// GIVEN a mock server with an existing route
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	dummy := mockserver.New("root", "master")
	mockserver.WithIPRoutes([]types.IPRoute{
		{
			ID:          "*1",
			DstAddress:  "0.0.0.0/0",
			Gateway:     "10.155.101.1",
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

	// WHEN an existing route is updated
	svc := &IPRouteService{c: c}
	updatedRoute := types.IPRouteAdd{
		DstAddress: "0.0.0.0/0",
		Gateway:    "10.155.101.2",
		Distance:   "2",
		Disabled:   "false",
	}

	got, err := svc.UpdateRoute(ctx, "*1", updatedRoute)

	// THEN it should return the updated route with the new values
	assert.NoError(t, err)
	assert.Equal(t, "0.0.0.0/0", got.DstAddress)
	assert.Equal(t, "10.155.101.2", got.Gateway)
	assert.Equal(t, "2", got.Distance)
	assert.Equal(t, "false", got.Disabled)
}
