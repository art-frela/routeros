# RouterOS Go Client

A Go client library for interacting with MikroTik RouterOS devices via the REST API.

[![Tests](https://github.com/art-frela/routeros/actions/workflows/go.yml/badge.svg)](https://github.com/art-frela/routeros/actions)
[![Coverage](https://codecov.io/gh/art-frela/routeros/branch/main/graph/badge.svg)](https://codecov.io/gh/art-frela/routeros)
[![Version](https://img.shields.io/github/v/tag/art-frela/routeros?label=version&sort=semver)](https://github.com/art-frela/routeros/tags)

## Features

This library provides a simple and intuitive interface for managing RouterOS devices. It includes:

- IP Firewall Address List management
- IP Address management
- Tool services (like ping)
- Easy configuration via environment variables
- Rate limiting to prevent overwhelming the RouterOS device
- Comprehensive test suite with mock server implementations

## Implemented API Endpoints

### IP Firewall Address List

- `GET /ip/firewall/address-list` - Find address list entries
- `PUT /ip/firewall/address-list` - Add new address list entries

### IP Addresses

- `GET /ip/address` - Get all IP addresses or find by ID
- `PUT /ip/address` - Add new IP address
- `DELETE /ip/address` - Remove IP address by ID
- `PATCH /ip/address` - Update IP address by ID

### Tools

- `POST /tool/ping` - Ping a host

## Installation

```bash
go get github.com/art-frela/routeros
```

## Usage

### Configuration

The client can be configured using environment variables:

```bash
export ROS_BASE_URL="http://192.168.88.1"
export ROS_USER="admin"
export ROS_PASSWORD="password"
export ROS_REQUEST_TIMEOUT="10s"
export ROS_PAUSE_BETWEEN_REQUESTS="100ms"
export ROS_BURST_REQ_COUNT="10"
```

### Creating a Client

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/art-frela/routeros"
)

func main() {
    // Create client configuration from environment variables
    cfg, err := routeros.NewClientConfigFromEnv("ROS")
    if err != nil {
        log.Fatal(err)
    }

    // Create client
    client, err := routeros.NewClient(*cfg)
    if err != nil {
        log.Fatal(err)
    }

    // Use the client
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Example: Get all IP addresses
    addresses, err := client.IPService.GetAddresses(ctx)
    if err != nil {
        log.Fatal(err)
    }

    for _, addr := range addresses {
        fmt.Printf("Address: %s, Interface: %s\n", addr.Address, addr.Interface)
    }
}
```

### IP Firewall Address List

```go
// Find address list entries
addresses, err := client.IPFirewallAddressListService.Find(ctx, "my-list", "192.168.1.100")
if err != nil {
    log.Fatal(err)
}

// Add new address to list
newItem := types.FirewallAddressListNewItem{
    Address: "192.168.1.100",
    List:    "my-list",
}

added, err := client.IPFirewallAddressListService.Add(ctx, newItem)
if err != nil {
    log.Fatal(err)
}
```

### IP Address Management

```go
// Get all IP addresses
addresses, err := client.IPService.GetAddresses(ctx)
if err != nil {
    log.Fatal(err)
}

// Add new IP address
newAddr := types.IPAddressAdd{
    Address:   "192.168.2.1/24",
    Interface: "ether2",
}

added, err := client.IPService.AddAddress(ctx, newAddr)
if err != nil {
    log.Fatal(err)
}

// Update IP address
updateAddr := types.IPAddressAdd{
    Address:   "192.168.2.2/24",
    Interface: "ether2",
    Disabled:  "true",
}

updated, err := client.IPService.UpdateAddress(ctx, added.ID, updateAddr)
if err != nil {
    log.Fatal(err)
}

// Remove IP address
err = client.IPService.RemoveAddress(ctx, added.ID)
if err != nil {
    log.Fatal(err)
}
```

### Tool Services

```go
// Ping a host
req := types.EchoRequest{
    Address: "8.8.8.8",
    Count:   3,
}

response, err := client.ToolService.Ping(ctx, req)
if err != nil {
    log.Fatal(err)
}

for _, echo := range response {
    fmt.Printf("Reply from %s: bytes=%s time=%s TTL=%s\n",
        echo.Host, *echo.Size, *echo.Time, *echo.TTL)
}
```

## Testing

The library includes a comprehensive test suite with mock server implementations:

```bash
go test -v ./...
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
