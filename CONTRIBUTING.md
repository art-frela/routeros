# Contributing to RouterOS Go Client

Thank you for your interest in contributing to the RouterOS Go client library!

## Development Workflow

### Requirements

- **Go** version 1.23 or higher
- **gofumpt** for code formatting
- **golangci-lint** for linting
- **testify** for testing assertions

### Code Quality

Before every commit, ensure:

```bash
# Format code
gofumpt -l -w .

# Run linter (uses .golangci.yml configuration)
golangci-lint run ./...

# Run tests with race detector
go test -race ./...
```

All tests must pass before submitting a Pull Request.

## API Stability

- **Pre-v1**: The public API may evolve. Breaking changes are possible between minor versions.
- **Additive-only**: Until v1.0, prefer adding new features rather than changing existing ones.
- **Deprecation**: Use `// Deprecated:` comments for obsolete APIs. Keep deprecated symbols for at least two releases to allow migration.

## Dependencies

### Runtime Dependencies

The project aims to keep runtime dependencies minimal. Current runtime dependencies:

- `github.com/kelseyhightower/envconfig` - Environment-based configuration
- `golang.org/x/time` - Rate limiting support

New runtime dependencies require justification in the Pull Request description.

### Test Dependencies

- `github.com/stretchr/testify` - Testing assertions, helpers and suites
- `github.com/testcontainers/testcontainers-go` v0.38.0 - Real RouterOS CHR containers for the integration tests (see [Integration Tests](#integration-tests) for the pin rationale)

Test dependencies are not part of the runtime footprint.

## Testing Conventions

### Test Structure

- Tests run **sequentially** (do not use `t.Parallel()` — tests rely on `t.Setenv`).
- Use **testify/assert** for all assertions.
- Include Gherkin-style comments: `// GIVEN`, `// WHEN`, `// THEN`.
- One `_test.go` file per service file.
- Use `pkg/mockserver` for HTTP mocking.

### Example Test Pattern

```go
func TestIPService_GetAddresses(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
    defer cancel()

    // GIVEN a mock server with pre-configured IP addresses
    dummy := mockserver.New("root", "master")
    mockserver.WithIPAddresses([]types.IPAddress{
        {ID: "*1", Address: "192.168.1.1/24", Interface: "ether1", Disabled: "false"},
    })(dummy)

    ts := httptest.NewServer(http.HandlerFunc(dummy.IPAddresses))
    defer ts.Close()

    t.Setenv("ROS_TEST_FIND_BASE_URL", ts.URL)
    t.Setenv("ROS_TEST_FIND_USER", "root")

    cfg, err := routeros.NewClientConfigFromEnv("ROS_TEST_FIND")
    assert.NoError(t, err)
    client, err := routeros.NewClient(*cfg)
    assert.NoError(t, err)

    // WHEN the IPService is called to get addresses
    // Public access: client.IPService (in-repo tests use &routeros.IPService{c: c} because they're in package routeros)
    ips := client.IPService
    got, err := ips.GetAddresses(ctx)

    // THEN it should return the expected addresses
    assert.NoError(t, err)
    assert.Len(t, got, 1)
    assert.Equal(t, "192.168.1.1/24", got[0].Address)
}
```

### Running Tests

```bash
# Run all tests
go test -v -race ./...

# With coverage
go test -v -race -coverprofile=coverage.out ./...
```

## Integration Tests

The integration suites exercise the client against a **real RouterOS CHR 7.23.3**
instead of mocks: the harness boots a CHR virtual router in QEMU inside an Alpine
container via [testcontainers-go](https://github.com/testcontainers/testcontainers-go),
waits until its REST API answers, provisions the `admin` password through the REST
API itself, and hands every suite a shared `*routeros.Client`.

A custom wrapper image is used instead of the official `mikrotik/chr` image
because the official one only works attached to a pre-existing host bridge
(started with `--network br1 --ipam-driver=none`) and crashes with a `tunl0`
error on standard Docker networks
([MikroTik forum](https://forum.mikrotik.com/t/181934),
[containerlab #3116](https://github.com/srl-labs/containerlab/issues/3116)).
The wrapper needs none of that: user-mode (SLIRP) networking, no `--privileged`,
no `NET_ADMIN`, no host bridge.

### Prerequisites

- **Docker** — that is all. No `sudo`, no extra daemons.
- **Linux**: `/dev/kvm` is attached automatically when present and writable,
  accelerating the boot; CI enables it with `sudo chmod 666 /dev/kvm`. Without
  KVM the router boots under TCG software emulation — slower, still green.
- **macOS (Apple Silicon)**: works out of the box — the harness builds the
  `aarch64` image variant and runs it under TCG emulation.

### Running

```bash
# Full integration suite (runs the unit tests too)
go test -tags=integration -timeout 25m -v ./...

# One domain suite only
go test -tags=integration -run 'TestIntegrationCHR_FirewallAddressList' -v -timeout 15m ./tests/integration/

# Plain unit run — integration tests are excluded unless -tags=integration is set
go test ./...
```

The **first run builds the wrapper image**, which downloads the RouterOS CHR
disk from download.mikrotik.com — expect a few minutes. Subsequent runs reuse
the built image (`KeepImage`), so only the container boot remains (roughly a
minute under TCG, seconds with KVM).

**No Docker running?** Every integration suite SKIPs with a clear message
(exit code 0) — it never fails the default `go test ./...` or CI unit runs.

### Harness Modes and Environment Variables

The harness picks the router source in this precedence:

1. **Bring your own device** — set `ROS_INTEGRATION_BASE_URL` (plus optional
   `ROS_INTEGRATION_USER` / `ROS_INTEGRATION_PASSWORD`; envconfig defaults are
   `root` / `master`): no container is ever started, and every integration
   suite — including the read-only `TestIntegrationBYO` smoke suite, which
   skips otherwise — runs against that device. Warning: the `TestIntegrationCHR_*`
   suites are not read-only; point them at a disposable router.
2. **Prebuilt image** — set `ROUTEROS_IT_IMAGE` to run a published image
   instead of building, e.g. the registry image
   `ghcr.io/art-frela/routeros-it-chr:7.23.3` (public package, anonymous pull,
   no `docker login`).
3. **Build locally** (default) — the harness builds `tests/integration/chr/` into
   `routeros-it-chr:<version>-<arch>` and reuses that image on later runs.

Additional knobs:

| Variable | Effect |
| --- | --- |
| `ROUTEROS_IT_VERSION` | RouterOS CHR version to build/boot (default `7.23.3`) |
| `ROUTEROS_IT_MEMORY` | Router memory in MiB, forwarded to the image's `QEMU_MEMORY` knob (default `512`) |
| `ROUTEROS_IT_CPUS` | Router vCPU count, forwarded to the image's `QEMU_CPUS` knob (default `2`) |

Example: `ROUTEROS_IT_IMAGE=ghcr.io/art-frela/routeros-it-chr:7.23.3 ROUTEROS_IT_MEMORY=1024 ROUTEROS_IT_CPUS=4 go test -tags=integration -timeout 25m -v ./...`

### Suite Layout and Conventions

Integration code lives in two places:

- `tests/integration/chr/Dockerfile` + `tests/integration/chr/entrypoint.sh` — the
  QEMU-wrapped CHR image: downloads and converts the CHR disk, forwards guest
  tcp/80 to container tcp/80, reads `QEMU_MEMORY` / `QEMU_CPUS` from the
  container environment.
- `tests/integration/` (build tag `//go:build integration`, package
  `integration`) — `main_test.go` holds the shared harness (container
  singleton, readiness wait, password provisioning, env handling, `BaseSuite`),
  plus one `<domain>_suite_test.go` per service.

Every domain suite embeds `BaseSuite` and is started by a plain runner
function:

```go
// FirewallAddressListSuite covers IPFirewallAddressListService (Find, Add)
// against the live RouterOS REST API bootstrapped by BaseSuite.
type FirewallAddressListSuite struct {
	BaseSuite
}

// TestIntegrationCHR_FirewallAddressList runs the FirewallAddressListSuite.
func TestIntegrationCHR_FirewallAddressList(t *testing.T) {
	suite.Run(t, new(FirewallAddressListSuite))
}

// TestAddAndFind verifies that Add creates a static entry and Find locates it.
func (s *FirewallAddressListSuite) TestAddAndFind() {
	// GIVEN a unique address-list entry
	listName := fmt.Sprintf("itlist-%d", time.Now().UnixNano())

	// WHEN the entry is added
	added, err := s.Client.IPFirewallAddressListService.Add(ctx, types.FirewallAddressListNewItem{
		Address: addr, List: listName,
	})

	// THEN the created item echoes the request
	s.Require().NoError(err)
	s.NotEmpty(added.ID)
}
```

Rules for integration tests:

- One `*_suite_test.go` per service, embedding `BaseSuite` (its `SetupSuite`
  bootstraps the shared router through `integrationClient(t)`).
- Runner named `TestIntegrationCHR_<Service>`, test methods `Test<Case>` —
  `go test` reports them as `TestIntegrationCHR_<Service>/<Case>`, and
  `-run 'TestIntegrationCHR_<Service>'` selects a single suite.
- Gherkin comments (`// GIVEN`, `// WHEN`, `// THEN`) and `s.Require()` /
  `s.Equal(...)` assertions, matching the unit-test style.
- NO `t.Parallel()` and NO `t.Setenv()` in integration tests — the router is
  a shared singleton and tests run sequentially.
- Make every resource unique (e.g. `time.Now().UnixNano()` suffixes) and remove
  it in `s.T().Cleanup(...)` so reruns never collide with leftovers. (The
  firewall address-list service exposes no Remove API; unique names make the
  accumulating entries harmless on the throwaway router.)
- Assert deterministic outcomes only: exact HTTP status ONLY where RouterOS
  behavior is documented — wrong password → exactly `401`, duplicate firewall
  address-list entry → exactly `400`; every other error case asserts
  `StatusCode >= 400`.
- Bound API calls with the shared `itTimeout` (30s) context — never the suite
  bootstrap, which may take minutes on a cold container.

### Adding a New Service Suite

Copy the closest existing `*_suite_test.go`: new file, embed `BaseSuite`, add
the runner, create unique resources and clean them up in `s.T().Cleanup(...)`.
Keep mock parity in mind: `pkg/mockserver` is expected to mirror live RouterOS
behavior, including validation errors — when an integration test uncovers a
divergence, fix the mock rather than special-casing the test.

### CI and the Published Image

- `.github/workflows/go.yml` runs the full integration suite on every push and
  PR with the exact command above. The job points `ROUTEROS_IT_IMAGE` at the
  prebuilt registry image (fast anonymous pull) and falls back to building
  `tests/integration/chr/` with the docker CLI when the image assets changed in the
  event (or when the registry image is not pullable yet), so a PR is never
  tested against a stale image. The CLI is used instead of the in-test
  testcontainers build because a BuildKit build through the raw Docker Engine
  API fails on plain `dockerd` daemons with `no active sessions` — only the
  docker CLI attaches the client-side BuildKit session. The same workaround
  applies locally if you ever hit that error: pre-build with
  `docker build -t routeros-it-chr:7.23.3-<arch> tests/integration/chr` and export
  `ROUTEROS_IT_IMAGE` with that tag.
- `.github/workflows/chr-image.yml` publishes the multi-arch
  (`linux/amd64` + `linux/arm64`) image to
  `ghcr.io/art-frela/routeros-it-chr:<version>` (and `:latest`) on pushes to
  `main` touching `tests/integration/chr/**`, plus `workflow_dispatch`. Note: the
  first publish of a new GHCR package is **private by default** — flip it to
  public once in the GitHub package settings so consumers can pull anonymously.

### Dependency Note

The harness depends on `github.com/testcontainers/testcontainers-go`, pinned
to **v0.38.0** — the last release line supporting the repository's Go 1.23
toolchain (v0.39.0 and newer require Go 1.24+). It is a **test-only**
dependency, imported exclusively from `//go:build integration` files, so it
adds nothing to the runtime footprint. Do not upgrade it past v0.38.x while
`go.mod` still declares `go 1.23`.

## Pull Request Process

1. Fork the repository and create a feature branch.
2. Follow the code quality checklist (format, lint, test).
3. Write clear commit messages and PR descriptions.
4. Ensure all CI checks pass.
5. Request review from maintainers.

## Code Style

- Follow standard Go conventions and idiomatic patterns.
- Use meaningful variable and function names.
- Keep functions focused and small.
- Document exported functions and types with comments.
