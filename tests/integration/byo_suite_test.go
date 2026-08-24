//go:build integration

package integration

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/art-frela/routeros"
	"github.com/art-frela/routeros/types"
)

// BYOSuite runs the legacy read-only smoke tests against a user-provided
// device configured via the ROS_INTEGRATION_* environment variables.
//
// It shadows BaseSuite.SetupSuite on purpose: no Docker health check and no
// CHR container — the BYO branch never starts a container.
type BYOSuite struct {
	BaseSuite
}

// SetupSuite builds the client from the ROS_INTEGRATION_* environment and
// skips the whole suite when no device is configured.
func (s *BYOSuite) SetupSuite() {
	// GIVEN a live device configured via ROS_INTEGRATION_BASE_URL
	if os.Getenv("ROS_INTEGRATION_BASE_URL") == "" {
		s.T().Skip("ROS_INTEGRATION_BASE_URL not set")
	}

	cfg, err := routeros.NewClientConfigFromEnv("ROS_INTEGRATION")
	s.Require().NoError(err)

	c, err := routeros.NewClient(*cfg)
	s.Require().NoError(err)

	s.Client = c
}

// TestIntegrationBYO runs the BYOSuite.
func TestIntegrationBYO(t *testing.T) {
	suite.Run(t, new(BYOSuite))
}

// TestIPService_GetAddresses lists IP addresses from a live device.
//
// Read-only: performs GET /rest/ip/address only, never mutates device state.
func (s *BYOSuite) TestIPService_GetAddresses() {
	ctx, cancel := context.WithTimeout(context.Background(), itTimeout)
	defer cancel()

	// WHEN the IPService lists all configured IP addresses
	addresses, err := s.Client.IPService.GetAddresses(ctx)

	// THEN it should return at least one address without error
	s.Require().NoError(err)
	s.NotNil(addresses)
	s.GreaterOrEqual(len(addresses), 1)
}

// TestIPRouteService_GetRoutes lists routes from a live device.
//
// Read-only: performs GET /rest/ip/route only, never mutates device state.
func (s *BYOSuite) TestIPRouteService_GetRoutes() {
	ctx, cancel := context.WithTimeout(context.Background(), itTimeout)
	defer cancel()

	// WHEN the IPRouteService lists all configured routes
	routes, err := s.Client.IPRouteService.GetRoutes(ctx)

	// THEN it should return the route list without error
	s.Require().NoError(err)
	s.NotNil(routes)
}

// TestToolService_Ping pings the device loopback once.
//
// Read-only: ping is a diagnostic action, it never mutates device state.
func (s *BYOSuite) TestToolService_Ping() {
	ctx, cancel := context.WithTimeout(context.Background(), itTimeout)
	defer cancel()

	// WHEN the ToolService pings 127.0.0.1 exactly once
	replies, err := s.Client.ToolService.Ping(ctx, types.EchoRequest{
		Address: "127.0.0.1",
		Count:   1,
	})

	// THEN it should return at least one reply element without error
	s.Require().NoError(err)
	s.NotNil(replies)
	s.GreaterOrEqual(len(replies), 1)
}
