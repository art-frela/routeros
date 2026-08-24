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

// FirewallAddressListSuite covers IPFirewallAddressListService (Find, Add)
// against the live RouterOS REST API bootstrapped by BaseSuite.
type FirewallAddressListSuite struct {
	BaseSuite
}

// TestIntegrationCHR_FirewallAddressList runs the FirewallAddressListSuite.
func TestIntegrationCHR_FirewallAddressList(t *testing.T) {
	suite.Run(t, new(FirewallAddressListSuite))
}

// TestAddAndFind verifies that Add creates a static entry and Find locates it
// through the combined, list-only and address-only query filters.
func (s *FirewallAddressListSuite) TestAddAndFind() {
	// The context bounds only the API calls below; the container bootstrap
	// in SetupSuite is not affected.
	ctx, cancel := context.WithTimeout(context.Background(), itTimeout)
	defer cancel()

	// GIVEN a unique firewall address-list entry: the list name is unique
	// per run (no Remove API exists — the service exposes only Find/Add —
	// so entries accumulate harmlessly on the throwaway CHR container and
	// unique names keep reruns collision-free). The optional fields are
	// omitempty since the Amendment-3 fix: a bare {Address, List} works
	// against live RouterOS.
	n := time.Now().UnixNano()%200 + 10
	listName := fmt.Sprintf("itlist-%d", time.Now().UnixNano())
	addr := fmt.Sprintf("192.168.%d.100", n)

	// WHEN the entry is added
	added, err := s.Client.IPFirewallAddressListService.Add(
		ctx,
		types.FirewallAddressListNewItem{
			Address: addr,
			List:    listName,
		},
	)

	// THEN the created item echoes the request with static-entry defaults
	s.Require().NoError(err)
	s.Require().NotNil(added)
	s.NotEmpty(added.ID)
	s.Equal(listName, added.List)
	s.Equal(addr, added.Address)
	s.Equal("false", added.Disabled)
	s.Equal("false", added.Dynamic)

	// WHEN Find filters by list AND address
	found, err := s.Client.IPFirewallAddressListService.Find(ctx, listName, addr)

	// THEN exactly the created entry matches
	s.Require().NoError(err)
	s.Require().Len(found, 1)
	s.Equal(added.ID, found[0].ID)

	// WHEN Find filters by list only
	byList, err := s.Client.IPFirewallAddressListService.Find(ctx, listName)

	// THEN at least the created entry is returned, all under that list
	s.Require().NoError(err)
	s.Require().NotEmpty(byList)
	for _, item := range byList {
		s.Equal(listName, item.List)
	}

	// WHEN Find filters by address only (empty list argument, variadic
	// address filter — the address-only query shape of Find)
	byAddress, err := s.Client.IPFirewallAddressListService.Find(ctx, "", addr)

	// THEN at least one entry carries exactly that address
	s.Require().NoError(err)
	s.Require().NotEmpty(byAddress)
	for _, item := range byAddress {
		s.Equal(addr, item.Address)
	}
}

// TestDuplicateRejected verifies RouterOS rejects adding the same
// {address, list} pair twice with 400 "failure: already have such entry"
// (mirrored by pkg/mockserver/ip_firewall.go:105-113).
func (s *FirewallAddressListSuite) TestDuplicateRejected() {
	// The context bounds only the API calls below.
	ctx, cancel := context.WithTimeout(context.Background(), itTimeout)
	defer cancel()

	// GIVEN a fresh unique {list, address} pair so reruns never collide
	// with entries from earlier runs
	n := time.Now().UnixNano()%200 + 10
	listName := fmt.Sprintf("itlist-%d", time.Now().UnixNano())
	addr := fmt.Sprintf("192.168.%d.101", n)
	item := types.FirewallAddressListNewItem{
		Address: addr,
		List:    listName,
	}

	// WHEN the pair is added twice
	_, err := s.Client.IPFirewallAddressListService.Add(ctx, item)
	s.Require().NoError(err)

	_, err = s.Client.IPFirewallAddressListService.Add(ctx, item)

	// THEN RouterOS rejects the duplicate with exactly 400 — the documented
	// "failure: already have such entry" rejection and the ONE sanctioned
	// exact-status assertion in the integration suites
	s.Require().Error(err)

	var rerr *routeros.ResponseError
	s.Require().True(errors.As(err, &rerr), "expected *routeros.ResponseError, got %T: %v", err, err)
	s.Require().Equal(400, rerr.StatusCode)
}
