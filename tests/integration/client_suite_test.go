//go:build integration

package integration

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/art-frela/routeros"
)

// ClientSuite covers Client-level behavior (authentication, error surface)
// against the live RouterOS REST API bootstrapped by BaseSuite.
type ClientSuite struct {
	BaseSuite
}

// TestIntegrationCHR_Client runs the ClientSuite.
func TestIntegrationCHR_Client(t *testing.T) {
	suite.Run(t, new(ClientSuite))
}

// TestWrongPassword verifies that a wrong password is rejected by the live
// RouterOS with 401 surfaced as *routeros.ResponseError.
func (s *ClientSuite) TestWrongPassword() {
	// GIVEN the bootstrapped harness client and a second client with a wrong password
	// The context bounds only the API calls below: the bootstrap in
	// SetupSuite can take minutes on a cold container and must not burn the
	// per-test budget.
	ctx, cancel := context.WithTimeout(context.Background(), itTimeout)
	defer cancel()

	wrong, err := routeros.NewClient(routeros.Config{
		BaseURL:  s.Client.BaseURL(),
		User:     chrAdminUser,
		Password: "definitely-wrong",
		// RequestTimeout must be non-zero: makeRequest layers a
		// context.WithTimeout(ctx, c.requestTimeout) over every call, and a
		// zero timeout expires the request before the 401 can arrive.
		RequestTimeout: itTimeout,
	})
	s.Require().NoError(err)

	// WHEN the wrong-password client lists IP addresses
	_, err = wrong.IPService.GetAddresses(ctx)

	// THEN RouterOS rejects the request with 401 as a *ResponseError
	s.Require().Error(err)

	var rerr *routeros.ResponseError
	s.Require().True(errors.As(err, &rerr), "expected *routeros.ResponseError, got %T: %v", err, err)
	s.Equal(http.StatusUnauthorized, rerr.StatusCode)
}
