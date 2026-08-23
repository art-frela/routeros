//go:build integration

package routeros

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/art-frela/routeros/types"
)

// skipIfIntegrationEnvUnset skips the test when no live device is configured.
// Integration tests are env-gated: they never run in default CI, only when
// ROS_INTEGRATION_BASE_URL points to a reachable RouterOS REST endpoint.
func skipIfIntegrationEnvUnset(t *testing.T) {
	t.Helper()

	if os.Getenv("ROS_INTEGRATION_BASE_URL") == "" {
		t.Skip("ROS_INTEGRATION_BASE_URL not set")
	}
}

// TestIntegration_IPService_GetAddresses lists IP addresses from a live device.
//
// Read-only: performs GET /rest/ip/address only, never mutates device state.
func TestIntegration_IPService_GetAddresses(t *testing.T) {
	skipIfIntegrationEnvUnset(t)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	// GIVEN a client configured from ROS_INTEGRATION_* environment variables
	cfg, err := NewClientConfigFromEnv("ROS_INTEGRATION")
	assert.NoError(t, err)

	c, err := NewClient(*cfg)
	assert.NoError(t, err)

	// WHEN the IPService lists all configured IP addresses
	addresses, err := c.IPService.GetAddresses(ctx)

	// THEN it should return at least one address without error
	assert.NoError(t, err)
	assert.NotNil(t, addresses)
	assert.GreaterOrEqual(t, len(addresses), 1)
}

// TestIntegration_IPRouteService_GetRoutes lists routes from a live device.
//
// Read-only: performs GET /rest/ip/route only, never mutates device state.
func TestIntegration_IPRouteService_GetRoutes(t *testing.T) {
	skipIfIntegrationEnvUnset(t)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	// GIVEN a client configured from ROS_INTEGRATION_* environment variables
	cfg, err := NewClientConfigFromEnv("ROS_INTEGRATION")
	assert.NoError(t, err)

	c, err := NewClient(*cfg)
	assert.NoError(t, err)

	// WHEN the IPRouteService lists all configured routes
	routes, err := c.IPRouteService.GetRoutes(ctx)

	// THEN it should return the route list without error
	assert.NoError(t, err)
	assert.NotNil(t, routes)
}

// TestIntegration_ToolService_Ping pings the device loopback once.
//
// Read-only: ping is a diagnostic action, it never mutates device state.
func TestIntegration_ToolService_Ping(t *testing.T) {
	skipIfIntegrationEnvUnset(t)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	// GIVEN a client configured from ROS_INTEGRATION_* environment variables
	cfg, err := NewClientConfigFromEnv("ROS_INTEGRATION")
	assert.NoError(t, err)

	c, err := NewClient(*cfg)
	assert.NoError(t, err)

	// WHEN the ToolService pings 127.0.0.1 exactly once
	replies, err := c.ToolService.Ping(ctx, types.EchoRequest{
		Address: "127.0.0.1",
		Count:   1,
	})

	// THEN it should return at least one reply element without error
	assert.NoError(t, err)
	assert.NotNil(t, replies)
	assert.GreaterOrEqual(t, len(replies), 1)
}
