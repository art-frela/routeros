//go:build integration

// Package integration hosts the testify/suite-based integration tests of the
// routeros client: a shared RouterOS CHR harness (this file) plus one
// *_suite_test.go per domain, modeled after rehub-service's tests/integration.
//
// The CHR container source is configurable: ROUTEROS_IT_IMAGE runs a
// prebuilt image instead of building the local test/integration context,
// and ROUTEROS_IT_MEMORY / ROUTEROS_IT_CPUS are forwarded to the image's
// QEMU_MEMORY / QEMU_CPUS entrypoint knobs (image defaults: 512 MiB, 2
// vCPUs).
package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/docker/docker/api/types/build"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/art-frela/routeros"
)

const (
	// chrAdminUser is the RouterOS user the harness provisions and uses.
	chrAdminUser = "admin"
	// chrPassword is the password provisioned for chrAdminUser after the
	// first CHR boot (NewClient rejects empty credentials by design).
	chrPassword = "master"
	// chrRouterOSVersion is the pinned RouterOS CHR version booted by the
	// harness; override it with the ROUTEROS_IT_VERSION env variable.
	chrRouterOSVersion = "7.23.3"

	// itTimeout bounds the per-test contexts handed to Client calls.
	itTimeout = 30 * time.Second

	// chrProvisionAttempts and chrProvisionRetry bound the password
	// provisioning loop (worst case ~10 requests + 9 sleeps).
	chrProvisionAttempts = 10
	chrProvisionRetry    = 3 * time.Second
)

// Singleton CHR container shared by every integration suite: it is built and
// booted once per `go test` process and intentionally kept alive for the
// whole run; the Ryuk reaper removes it after the process exits (KeepImage
// preserves the built image for the next run).
var (
	chrOnce   sync.Once
	chrClient *routeros.Client
	chrErr    error
)

// integrationClient returns a Client connected to a live RouterOS REST API.
//
// Container source precedence: BYO device (ROS_INTEGRATION_BASE_URL set —
// client built from the ROS_INTEGRATION_* environment, no container
// started) > prebuilt image (ROUTEROS_IT_IMAGE, see startCHR) > build from
// the local test/integration context. In the container modes the singleton
// RouterOS CHR is built/pulled (first run only), booted, the admin
// password is provisioned, and the shared client is returned. The test is
// skipped with a clear message when Docker is missing or unhealthy.
func integrationClient(t *testing.T) *routeros.Client {
	t.Helper()

	testcontainers.SkipIfProviderIsNotHealthy(t)

	// BYO device: mirror the BYOSuite env semantics, no container.
	if os.Getenv("ROS_INTEGRATION_BASE_URL") != "" {
		cfg, err := routeros.NewClientConfigFromEnv("ROS_INTEGRATION")
		if err != nil {
			t.Fatalf("load ROS_INTEGRATION config: %v", err)
		}

		c, err := routeros.NewClient(*cfg)
		if err != nil {
			t.Fatalf("create ROS_INTEGRATION client: %v", err)
		}

		return c
	}

	chrOnce.Do(startCHR)
	if chrErr != nil {
		t.Fatalf("CHR bootstrap failed: %v", chrErr)
	}

	return chrClient
}

// startCHR creates the singleton CHR container — running the prebuilt image
// referenced by ROUTEROS_IT_IMAGE when set, building the local
// test/integration context otherwise — then boots it, provisions the admin
// password and stores the shared client.
//
// It MUST NOT call t.Skip/t.Fatalf or anything that runtime.Goexits:
// sync.Once.Do is marked done even on Goexit, so bailing out here would hand
// every later test a nil chrClient. Failures are recorded in chrErr, which
// integrationClient reports outside the Once.
func startCHR() {
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Minute)
	defer cancel()

	ver := chrVersion()
	arch := chrArch()

	// Container source precedence (the BYO device, ROS_INTEGRATION_BASE_URL,
	// sits above both and is handled in integrationClient): prebuilt image
	// (ROUTEROS_IT_IMAGE) > build from the local test/integration context.
	// ROUTEROS_IT_MEMORY and ROUTEROS_IT_CPUS are forwarded to the image's
	// QEMU_MEMORY/QEMU_CPUS entrypoint knobs via chrQemuEnv (unset knobs
	// fall back to the image defaults).
	req := testcontainers.ContainerRequest{
		Env:          chrQemuEnv(),
		ExposedPorts: []string{"80/tcp"},
		// The wait probe hits the REST endpoint itself: a fresh CHR answers
		// 401 there, which already proves REST is serving (no WebFig "/" 200
		// assumption). WithForcedIPv4LocalHost dodges the Docker-Desktop
		// ::1 localhost pitfall on macOS.
		WaitingFor: wait.ForHTTP("/rest/system/resource").
			WithPort("80/tcp").
			WithForcedIPv4LocalHost().
			WithStatusCodeMatcher(func(code int) bool { return code >= 200 && code < 500 }).
			WithStartupTimeout(10 * time.Minute),
		// /dev/kvm acceleration when the host has it (Linux); nil elsewhere
		// (macOS) — the entrypoint then falls back to TCG emulation.
		HostConfigModifier: kvmHostConfigModifier(),
	}

	if imageRef := chrImageRef(); imageRef != "" {
		// Prebuilt image: run the referenced image ref directly instead of
		// building test/integration (fast path, e.g. the published
		// multi-arch registry image).
		req.Image = imageRef
	} else {
		// No image ref: build from the local test/integration context so a
		// fresh clone runs without any registry access.
		req.FromDockerfile = testcontainers.FromDockerfile{
			// Relative to the test binary CWD (tests/integration), hence
			// two levels up to the repo's test/integration assets.
			Context:    "../../test/integration",
			Dockerfile: "Dockerfile",
			// Fixed repo/tag + KeepImage avoid the default UUID-tag rebuild
			// churn; Docker layer cache makes reruns cheap.
			Repo:      "routeros-it-chr",
			Tag:       ver + "-" + arch,
			KeepImage: true,
			BuildArgs: map[string]*string{
				"ROUTEROS_VERSION": &ver,
				"ROUTEROS_ARCH":    &arch,
			},
			// Recent Docker daemons build without BuildKit by default on
			// the API path and reject the Dockerfile's COPY --chmod.
			BuildOptionsModifier: func(opts *build.ImageBuildOptions) {
				opts.Version = build.BuilderBuildKit
			},
		}
	}

	ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		chrErr = fmt.Errorf("start CHR container: %w", err)
		return
	}

	host, err := ctr.Host(ctx)
	if err != nil {
		chrErr = fmt.Errorf("resolve CHR host: %w", err)
		return
	}

	port, err := ctr.MappedPort(ctx, "80/tcp")
	if err != nil {
		chrErr = fmt.Errorf("resolve CHR mapped port: %w", err)
		return
	}

	baseURL := "http://" + net.JoinHostPort(host, port.Port())

	if err := provisionCHRPassword(ctx, baseURL); err != nil {
		chrErr = fmt.Errorf("provision CHR admin password: %w", err)
		return
	}

	chrClient, chrErr = routeros.NewClient(routeros.Config{
		BaseURL:        baseURL,
		User:           chrAdminUser,
		Password:       chrPassword,
		RequestTimeout: itTimeout,
	})
}

// chrImageRef returns the prebuilt CHR image reference to run instead of
// building the local test/integration context (ROUTEROS_IT_IMAGE), or ""
// when unset.
func chrImageRef() string {
	return os.Getenv("ROUTEROS_IT_IMAGE")
}

// chrQemuEnv forwards ROUTEROS_IT_MEMORY / ROUTEROS_IT_CPUS onto the image
// contract knobs QEMU_MEMORY / QEMU_CPUS. Unset variables are omitted so the
// image defaults (512 MiB, 2 vCPUs) apply.
func chrQemuEnv() map[string]string {
	env := make(map[string]string)

	if v := os.Getenv("ROUTEROS_IT_MEMORY"); v != "" {
		env["QEMU_MEMORY"] = v
	}

	if v := os.Getenv("ROUTEROS_IT_CPUS"); v != "" {
		env["QEMU_CPUS"] = v
	}

	return env
}

// chrVersion returns the RouterOS CHR version to boot: ROUTEROS_IT_VERSION
// when set, chrRouterOSVersion otherwise.
func chrVersion() string {
	if v := os.Getenv("ROUTEROS_IT_VERSION"); v != "" {
		return v
	}

	return chrRouterOSVersion
}

// chrArch maps the test binary architecture onto the Dockerfile ROUTEROS_ARCH
// build argument.
func chrArch() string {
	switch runtime.GOARCH {
	case "arm64":
		return "aarch64"
	default:
		return "x86_64"
	}
}

// kvmHostConfigModifier returns a HostConfigModifier attaching /dev/kvm to
// the container when the host exposes it, or nil when it does not (macOS and
// other TCG-only hosts). DeviceMapping through the Docker HostConfig is the
// only way to attach host devices in testcontainers-go v0.38.0.
func kvmHostConfigModifier() func(*container.HostConfig) {
	if _, err := os.Stat("/dev/kvm"); err != nil {
		return nil
	}

	return func(hc *container.HostConfig) {
		hc.Devices = []container.DeviceMapping{
			{PathOnHost: "/dev/kvm", PathInContainer: "/dev/kvm"},
		}
	}
}

// provisionCHRPassword sets the admin password on a freshly booted CHR via
// plain net/http and verifies the new credentials.
//
// Per attempt the verify GET runs first (idempotence: once the password is
// set, the empty password no longer authenticates, so a racing verify in an
// earlier attempt must not doom the loop), then the provisioning POST runs
// and is verified again. Provisioning uses POST /rest/execute with the
// script "/user set admin password=master" authenticated as the fresh-CHR
// empty-password admin (accepted on REST despite expired:true); when
// /rest/execute answers 4xx, the loop switches to the also-documented
// POST /rest/password endpoint.
func provisionCHRPassword(ctx context.Context, baseURL string) error {
	hc := &http.Client{Timeout: 10 * time.Second}

	usePasswordEndpoint := false
	var lastErr error

	for attempt := 1; attempt <= chrProvisionAttempts; attempt++ {
		if err := verifyCHRPassword(ctx, hc, baseURL); err == nil {
			return nil
		}

		var err error
		if usePasswordEndpoint {
			err = postCHRPassword(ctx, hc, baseURL)
		} else {
			var switchToPassword bool
			switchToPassword, err = postCHRExecuteScript(ctx, hc, baseURL)
			if switchToPassword {
				usePasswordEndpoint = true
			}
		}

		lastErr = err
		if err == nil {
			lastErr = verifyCHRPassword(ctx, hc, baseURL)
			if lastErr == nil {
				return nil
			}
		}

		if attempt == chrProvisionAttempts {
			break
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("provisioning interrupted: %w", ctx.Err())
		case <-time.After(chrProvisionRetry):
		}
	}

	return fmt.Errorf("password not verified after %d attempts: %w", chrProvisionAttempts, lastErr)
}

// postCHRExecuteScript runs one provisioning POST to /rest/execute with the
// script that sets the admin password, authenticated as the empty-password
// admin of a fresh CHR. It reports whether the endpoint answered 4xx, in
// which case the caller must switch to POST /rest/password.
func postCHRExecuteScript(ctx context.Context, hc *http.Client, baseURL string) (switchToPassword bool, err error) {
	payload := map[string]string{"script": "/user set " + chrAdminUser + " password=" + chrPassword}

	_, err = chrPost(ctx, hc, baseURL+"/rest/execute", payload, chrAdminUser, "")
	if err == nil {
		return false, nil
	}

	var rerr *routeros.ResponseError
	if errors.As(err, &rerr) && rerr.StatusCode >= 400 && rerr.StatusCode < 500 {
		return true, nil
	}

	return false, err
}

// postCHRPassword runs one provisioning POST to the /rest/password endpoint,
// the documented fallback for setting a fresh admin password.
func postCHRPassword(ctx context.Context, hc *http.Client, baseURL string) error {
	payload := map[string]string{"old-password": "", "new-password": chrPassword}

	_, err := chrPost(ctx, hc, baseURL+"/rest/password", payload, chrAdminUser, "")
	return err
}

// verifyCHRPassword checks that the provisioned credentials authenticate:
// GET /rest/system/resource with admin:chrPassword must answer 200.
func verifyCHRPassword(ctx context.Context, hc *http.Client, baseURL string) error {
	url := baseURL + "/rest/system/resource"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("build verify request: %w", err)
	}

	req.SetBasicAuth(chrAdminUser, chrPassword)

	res, err := hc.Do(req)
	if err != nil {
		return fmt.Errorf("GET %s: %w", url, err)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("verify GET %s: status_code: %d", url, res.StatusCode)
	}

	return nil
}

// chrPost sends a JSON POST with basic auth and returns the HTTP status; a
// status >= 400 is returned as a *routeros.ResponseError, mirroring
// makeRequest.
func chrPost(ctx context.Context, hc *http.Client, url string, payload any, user, password string) (int, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("marshal %s body: %w", url, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("build %s request: %w", url, err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(user, password)

	res, err := hc.Do(req)
	if err != nil {
		return 0, fmt.Errorf("POST %s: %w", url, err)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode >= http.StatusBadRequest {
		// Best-effort body capture for diagnostics only.
		respBody, _ := io.ReadAll(io.LimitReader(res.Body, 1024))
		return res.StatusCode, &routeros.ResponseError{StatusCode: res.StatusCode, Body: string(respBody)}
	}

	return res.StatusCode, nil
}

// BaseSuite is embedded by every integration suite; it exposes the shared
// live-router client.
type BaseSuite struct {
	suite.Suite
	Client *routeros.Client
}

// SetupSuite bootstraps the shared CHR container (or BYO device).
//
// There is deliberately NO TestMain: the reference layout bootstraps in
// TestMain with log.Fatalf, but that would FAIL the run instead of SKIPping
// when Docker is missing — this suite's SetupSuite keeps the plan-required
// "no Docker -> SKIP with clear message" semantics.
func (s *BaseSuite) SetupSuite() {
	s.Client = integrationClient(s.T())
}
