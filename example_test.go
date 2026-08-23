// Package routeros_test contains compile-only godoc examples for the public
// API of the routeros package.
//
// None of the examples below declare expected output, so the Go testing
// framework compiles them but never executes them — every call would require
// a live RouterOS device answering at the configured BaseURL.
package routeros_test

import (
	"context"
	"fmt"
	"time"

	"github.com/art-frela/routeros"
	"github.com/art-frela/routeros/types"
)

// ExampleNewClient builds a client from an explicitly constructed Config.
//
// RequestTimeout is set explicitly because the zero value would expire every
// request context immediately at runtime.
func ExampleNewClient() {
	cfg := routeros.Config{
		BaseURL:        "http://192.168.88.1",
		RequestTimeout: 10 * time.Second,
		User:           "admin",
		Password:       "master",
	}

	client, err := routeros.NewClient(cfg)
	if err != nil {
		fmt.Println("client error:", err)
		return
	}

	fmt.Println("client base URL:", client.BaseURL())
}

// ExampleNewClientConfigFromEnv loads the client configuration from
// environment variables prefixed with ROS_ (ROS_BASE_URL, ROS_USER and
// ROS_PASSWORD — see the Config documentation for the full list).
func ExampleNewClientConfigFromEnv() {
	cfg, err := routeros.NewClientConfigFromEnv("ROS")
	if err != nil {
		fmt.Println("config error:", err)
		return
	}

	client, err := routeros.NewClient(*cfg)
	if err != nil {
		fmt.Println("client error:", err)
		return
	}

	fmt.Println("client base URL:", client.BaseURL())
}

// ExampleIPService_GetAddresses lists all IP addresses configured on the
// device and prints each address with its interface.
func ExampleIPService_GetAddresses() {
	cfg := routeros.Config{
		BaseURL:        "http://192.168.88.1",
		RequestTimeout: 10 * time.Second,
		User:           "admin",
		Password:       "master",
	}

	client, err := routeros.NewClient(cfg)
	if err != nil {
		fmt.Println("client error:", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	addresses, err := client.IPService.GetAddresses(ctx)
	if err != nil {
		fmt.Println("request error:", err)
		return
	}

	for _, addr := range addresses {
		fmt.Printf("%s on %s\n", addr.Address, addr.Interface)
	}
}

// ExampleIPFirewallAddressListService_Find retrieves firewall address list
// entries filtered by list name and address.
func ExampleIPFirewallAddressListService_Find() {
	cfg := routeros.Config{
		BaseURL:        "http://192.168.88.1",
		RequestTimeout: 10 * time.Second,
		User:           "admin",
		Password:       "master",
	}

	client, err := routeros.NewClient(cfg)
	if err != nil {
		fmt.Println("client error:", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	entries, err := client.IPFirewallAddressListService.Find(ctx, "blacklist", "192.168.1.100")
	if err != nil {
		fmt.Println("request error:", err)
		return
	}

	for _, entry := range entries {
		fmt.Printf("%s list=%s disabled=%s\n", entry.Address, entry.List, entry.Disabled)
	}
}

// ExampleIPFirewallAddressListService_Add adds a new entry to a firewall
// address list and prints the stored item with its assigned ID.
func ExampleIPFirewallAddressListService_Add() {
	cfg := routeros.Config{
		BaseURL:        "http://192.168.88.1",
		RequestTimeout: 10 * time.Second,
		User:           "admin",
		Password:       "master",
	}

	client, err := routeros.NewClient(cfg)
	if err != nil {
		fmt.Println("client error:", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	item := types.FirewallAddressListNewItem{
		Address: "192.168.1.100",
		List:    "blacklist",
		Comment: "blocked by example",
	}

	added, err := client.IPFirewallAddressListService.Add(ctx, item)
	if err != nil {
		fmt.Println("request error:", err)
		return
	}

	fmt.Printf("added %s to list %s with id %s\n", added.Address, added.List, added.ID)
}

// ExampleToolService_Ping pings a host from the device and prints one line
// per reply; timed-out replies carry a status instead of RTT fields.
func ExampleToolService_Ping() {
	cfg := routeros.Config{
		BaseURL:        "http://192.168.88.1",
		RequestTimeout: 10 * time.Second,
		User:           "admin",
		Password:       "master",
	}

	client, err := routeros.NewClient(cfg)
	if err != nil {
		fmt.Println("client error:", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := types.EchoRequest{
		Address:  "8.8.8.8",
		Count:    3,
		Interval: 1,
	}

	replies, err := client.ToolService.Ping(ctx, req)
	if err != nil {
		fmt.Println("request error:", err)
		return
	}

	for _, echo := range replies {
		if echo.Time != nil {
			fmt.Printf("reply from %s: seq=%s time=%s\n", echo.Host, echo.Seq, *echo.Time)
			continue
		}

		fmt.Printf("reply from %s: seq=%s status=%s\n", echo.Host, echo.Seq, *echo.Status)
	}
}
