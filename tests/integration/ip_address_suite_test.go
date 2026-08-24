//go:build integration

package integration

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/art-frela/routeros"
	"github.com/art-frela/routeros/types"
)

// IPAddressSuite covers IPService (GetAddresses, GetAddressByID, AddAddress,
// UpdateAddress, RemoveAddress) against the live RouterOS REST API
// bootstrapped by BaseSuite.
type IPAddressSuite struct {
	BaseSuite
}

// TestIntegrationCHR_IPService runs the IPAddressSuite.
func TestIntegrationCHR_IPService(t *testing.T) {
	suite.Run(t, new(IPAddressSuite))
}

// TestCRUD verifies the full address lifecycle: Add, list, Get by id,
// Update (path-form id since the Amendment-3 fix), Remove, and the
// zero-struct answer of GetAddressByID for well-formed but absent ids.
func (s *IPAddressSuite) TestCRUD() {
	// The context bounds only the API calls below; the container bootstrap
	// in SetupSuite is not affected.
	ctx, cancel := context.WithTimeout(context.Background(), itTimeout)
	defer cancel()

	// GIVEN a unique address pair on the always-present ether1 interface
	n := time.Now().UnixNano()%200 + 10
	addr := fmt.Sprintf("192.168.%d.1/24", n)
	nextAddr := fmt.Sprintf("192.168.%d.2/24", n)

	// WHEN the address is added
	added, err := s.Client.IPService.AddAddress(ctx, types.IPAddressAdd{
		Address:   addr,
		Interface: "ether1",
	})

	// THEN the created entry echoes the request and carries an id
	s.Require().NoError(err)
	s.NotEmpty(added.ID)
	s.Equal(addr, added.Address)

	// Safety net: remove the address even when a later step fails; a fresh
	// context is needed because the test context is cancelled by then.
	s.T().Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), itTimeout)
		defer cleanupCancel()

		_ = s.Client.IPService.RemoveAddress(cleanupCtx, added.ID)
	})

	// WHEN all addresses are listed
	list, err := s.Client.IPService.GetAddresses(ctx)

	// THEN the created entry is present
	s.Require().NoError(err)
	found := false

	for _, ip := range list {
		if ip.ID == added.ID {
			found = true
			break
		}
	}

	s.True(found, "added address %s (%s) not found in GetAddresses result", added.ID, addr)

	// WHEN the address is fetched by its id
	byID, err := s.Client.IPService.GetAddressByID(ctx, added.ID)

	// THEN it reports the created address
	s.Require().NoError(err)
	s.Equal(addr, byID.Address)

	// WHEN the address is updated to the second address of the /24
	updated, err := s.Client.IPService.UpdateAddress(ctx, added.ID, types.IPAddressAdd{
		Address:   nextAddr,
		Interface: "ether1",
	})

	// THEN the update succeeds and echoes the new address
	s.Require().NoError(err)
	s.Equal(nextAddr, updated.Address)

	// WHEN the address is removed
	err = s.Client.IPService.RemoveAddress(ctx, added.ID)

	// THEN the removal succeeds
	s.NoError(err)

	// WHEN the removed address is fetched by its id again
	after, err := s.Client.IPService.GetAddressByID(ctx, added.ID)

	// THEN RouterOS answers 200 with an empty list, surfaced as the zero
	// struct (empty ID) rather than an error
	s.Require().NoError(err)
	s.Empty(after.ID)

	// WHEN a well-formed but nonexistent id is fetched
	// (worker-verified against live CHR: 200 [] -> zero struct)
	defaced, err := s.Client.IPService.GetAddressByID(ctx, "*DEFACE")

	// THEN it also yields the zero struct without an error
	s.Require().NoError(err)
	s.Empty(defaced.ID)
}

// TestInvalidInterface verifies that adding an address on a nonexistent
// interface is rejected as a *routeros.ResponseError with a client-error
// status (the exact code/wording varies across RouterOS versions).
func (s *IPAddressSuite) TestInvalidInterface() {
	// The context bounds only the API calls below.
	ctx, cancel := context.WithTimeout(context.Background(), itTimeout)
	defer cancel()

	// GIVEN the bootstrapped client and an address on a bogus interface

	// WHEN the address is added
	_, err := s.Client.IPService.AddAddress(ctx, types.IPAddressAdd{
		Address:   "192.168.250.1/24",
		Interface: "no-such-iface",
	})

	// THEN RouterOS rejects it as a ResponseError with status >= 400
	s.Require().Error(err)

	var rerr *routeros.ResponseError
	s.Require().True(errors.As(err, &rerr), "expected *routeros.ResponseError, got %T: %v", err, err)
	s.GreaterOrEqual(rerr.StatusCode, 400)
}
