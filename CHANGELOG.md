# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Injectable `*http.Client` via `Config.HTTPClient` for custom transport, TLS, and proxy control
- Exported typed errors: `ErrMissBaseURL`, `ErrMissUserOrPass`, `ResponseError`
- Package documentation (`doc.go`) with godoc examples
- `CHANGELOG.md` backfilled from git history
- `.golangci.yml` linter configuration with CI lint/race/Go-matrix parity
- Versioning and API-stability policy (README, CONTRIBUTING)
- Dependency policy and env-gated integration tests behind build tag
- `/ip/address` endpoint with full CRUD methods (Get, GetByID, Add, Remove, Update)
- `/ip/route` endpoint with full CRUD methods

### Changed

- Decode errors now wrapped with `%w` for error chain support

## [v0.0.3] - 2025-05-10

### Added

- `BaseURL` method on `Client`

## [v0.0.2] - 2025-05-09

### Added

- `ToolService` with ping method (`/tool/ping`)

## [v0.0.1] - 2025-04-12

### Added

- `IPFirewallAddressListService` with Find and Add methods (`/ip/firewall/address-list`)
- REST client with basic auth, rate limiting, and generic `makeRequest[T]` engine
- Mock server for testing
- README with usage examples

Format: Keep a Changelog, semantics: SemVer
