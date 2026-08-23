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

- `github.com/stretchr/testify` - Testing assertions and helpers

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
