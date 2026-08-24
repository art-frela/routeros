//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/art-frela/routeros/types"
)

// ToolSuite covers the ToolService (ping) against the live RouterOS REST API
// bootstrapped by BaseSuite.
type ToolSuite struct {
	BaseSuite
}

// TestIntegrationCHR_ToolService runs the ToolSuite.
func TestIntegrationCHR_ToolService(t *testing.T) {
	suite.Run(t, new(ToolSuite))
}

// TestPingLoopback verifies that pinging the RouterOS loopback address
// returns successful echo replies.
//
// Read-only: ping is a diagnostic action, it never mutates device state.
// The loopback address is chosen deliberately: it answers from the router
// itself, so no external egress (non-deterministic in CI) is involved.
func (s *ToolSuite) TestPingLoopback() {
	// GIVEN the bootstrapped harness client and a bounded context
	// (30s covers a count=2 ping with 0.2s interval easily)
	ctx, cancel := context.WithTimeout(context.Background(), itTimeout)
	defer cancel()

	// WHEN the ToolService pings the router's own loopback address twice
	replies, err := s.Client.ToolService.Ping(ctx, types.EchoRequest{
		Address:  "127.0.0.1",
		Count:    2,
		Interval: 0.2,
	})

	// THEN the ping succeeds and returns a successful echo reply
	// (len is NOT asserted equal to Count: real RouterOS may batch
	// per-reply rows with a trailing summary row).
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(len(replies), 1)
	s.Require().Equal("127.0.0.1", replies[0].Host)
	// Success replies omit the status field entirely.
	s.Require().Nil(replies[0].Status)
	// A successful reply carries timing evidence: TTL or Time must be
	// present (OR-shaped so field-presence nuances on real RouterOS
	// don't flake; loopback always replies).
	s.Require().True(replies[0].TTL != nil || replies[0].Time != nil,
		"expected TTL or Time in the first reply, got %+v", replies[0])
}
