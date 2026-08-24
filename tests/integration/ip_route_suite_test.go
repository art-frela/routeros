//go:build integration

package integration

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/art-frela/routeros"
	"github.com/art-frela/routeros/types"
)

// IPRouteSuite covers IPRouteService (GetRoutes, GetRouteByID, AddRoute,
// UpdateRoute, RemoveRoute) against the live RouterOS REST API bootstrapped
// by BaseSuite.
type IPRouteSuite struct {
	BaseSuite
}

// TestIntegrationCHR_IPRouteService runs the IPRouteSuite.
func TestIntegrationCHR_IPRouteService(t *testing.T) {
	suite.Run(t, new(IPRouteSuite))
}

// TestCRUD verifies the full static-route lifecycle on top of a
// prerequisite address: the connected route it installs, Add, Get by id,
// Update (path-form id since the Amendment-3 fix), Remove, and the
// zero-struct answer of GetRouteByID for a removed id.
func (s *IPRouteSuite) TestCRUD() {
	// The context bounds only the API calls below (including the
	// connected-route poll: worst case 15 attempts x 500 ms + request
	// times, well within itTimeout); the container bootstrap in
	// SetupSuite is not affected.
	ctx, cancel := context.WithTimeout(context.Background(), itTimeout)
	defer cancel()

	// GIVEN a unique prerequisite address (the static route's gateway)
	n := time.Now().UnixNano()%200 + 10
	gw := fmt.Sprintf("192.168.%d.1", n)
	ipAdd, err := s.Client.IPService.AddAddress(ctx, types.IPAddressAdd{
		Address:   gw + "/24",
		Interface: "ether1",
	})
	s.Require().NoError(err)

	// Safety net: remove the prerequisite address even when a later step
	// fails; a fresh context is needed because the test context is
	// cancelled by then. Registered FIRST so it runs AFTER the route
	// cleanup below (cleanups are LIFO and the route's gateway must stay
	// reachable until the route is gone).
	s.T().Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), itTimeout)
		defer cleanupCancel()

		_ = s.Client.IPService.RemoveAddress(cleanupCtx, ipAdd.ID)
	})

	// WHEN routes are listed (connected routes install ~1-3s after the
	// address add on live CHR, so poll before asserting)
	connected := fmt.Sprintf("192.168.%d.0/24", n)
	s.Require().True(
		s.waitForRouteWithDstPrefix(ctx, connected),
		"connected route for %s not installed within the poll budget", connected,
	)

	routes, err := s.Client.IPRouteService.GetRoutes(ctx)

	// THEN the table is non-empty and carries the connected route
	s.Require().NoError(err)
	s.NotEmpty(routes)

	found := false

	for _, r := range routes {
		if strings.HasPrefix(r.DstAddress, connected) {
			found = true
			break
		}
	}

	s.True(found, "no route with dst-address prefix %s in GetRoutes result", connected)

	// WHEN a static route inside the 10.x test range is added
	dst := fmt.Sprintf("10.%d.0.0/16", n)
	route, err := s.Client.IPRouteService.AddRoute(ctx, types.IPRouteAdd{
		DstAddress: dst,
		Gateway:    gw,
	})

	// THEN the created route carries an id
	s.Require().NoError(err)
	s.NotEmpty(route.ID)

	// Safety net for the route (LIFO: runs before the address cleanup)
	s.T().Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), itTimeout)
		defer cleanupCancel()

		_ = s.Client.IPRouteService.RemoveRoute(cleanupCtx, route.ID)
	})

	// WHEN the route is fetched by its id
	byID, err := s.Client.IPRouteService.GetRouteByID(ctx, route.ID)

	// THEN it echoes the created destination
	s.Require().NoError(err)
	s.Equal(dst, byID.DstAddress)

	// WHEN the route is updated with a comment
	updated, err := s.Client.IPRouteService.UpdateRoute(ctx, route.ID, types.IPRouteAdd{
		DstAddress: dst,
		Gateway:    gw,
		Comment:    "it-updated",
	})

	// THEN the update succeeds and the comment is persisted
	s.Require().NoError(err)
	s.Equal("it-updated", updated.Comment)

	// WHEN the route is removed
	err = s.Client.IPRouteService.RemoveRoute(ctx, route.ID)

	// THEN the removal succeeds
	s.NoError(err)

	// WHEN the removed route is fetched by its id again
	after, err := s.Client.IPRouteService.GetRouteByID(ctx, route.ID)

	// THEN the lookup yields the zero struct (empty ID) without an error
	s.Require().NoError(err)
	s.Empty(after.ID)
}

// TestFailures verifies the two deterministic rejection paths: removing a
// nonexistent id and adding a route without a dst-address. Since the
// Amendment-3 path-form fix, RouterOS answers 404 for unknown ids — still
// asserted as >= 400, the stable contract.
func (s *IPRouteSuite) TestFailures() {
	// The context bounds only the API calls below.
	ctx, cancel := context.WithTimeout(context.Background(), itTimeout)
	defer cancel()

	// GIVEN the bootstrapped client

	// WHEN a well-formed but nonexistent route id is removed
	err := s.Client.IPRouteService.RemoveRoute(ctx, "*DEFACE")

	// THEN RouterOS rejects it as a ResponseError with status >= 400
	s.Require().Error(err)

	var rerr *routeros.ResponseError
	s.Require().True(errors.As(err, &rerr), "expected *routeros.ResponseError, got %T: %v", err, err)
	s.GreaterOrEqual(rerr.StatusCode, 400)

	// WHEN a route is added without its required dst-address
	n := time.Now().UnixNano()%200 + 10
	_, err = s.Client.IPRouteService.AddRoute(ctx, types.IPRouteAdd{
		Gateway: fmt.Sprintf("192.168.%d.1", n),
	})

	// THEN RouterOS rejects it as a ResponseError with status >= 400
	s.Require().Error(err)

	rerr = nil
	s.Require().True(errors.As(err, &rerr), "expected *routeros.ResponseError, got %T: %v", err, err)
	s.GreaterOrEqual(rerr.StatusCode, 400)
}

// waitForRouteWithDstPrefix polls GetRoutes until at least one route's
// dst-address starts with prefix, or the attempt budget (15 x 500 ms) is
// exhausted. Connected routes need ~1-3s to appear after the address add on
// live CHR, so the CRUD test must poll instead of asserting immediately.
func (s *IPRouteSuite) waitForRouteWithDstPrefix(ctx context.Context, prefix string) bool {
	for attempt := 0; attempt < 15; attempt++ {
		routes, err := s.Client.IPRouteService.GetRoutes(ctx)
		if err == nil {
			for _, r := range routes {
				if strings.HasPrefix(r.DstAddress, prefix) {
					return true
				}
			}
		}

		select {
		case <-ctx.Done():
			return false
		case <-time.After(500 * time.Millisecond):
		}
	}

	return false
}
