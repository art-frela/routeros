// Package routeros implements a client for the MikroTik RouterOS REST API.
//
// The entry point is a [Client], created from a [Config] (loadable from
// environment variables with [NewClientConfigFromEnv]) that carries the base
// URL, credentials, request timeout, and an optional custom *http.Client.
// The client groups typed services covering IP firewall address lists, IP
// addresses, IP routes, and the ping tool.
//
// Every request passes through a rate.Limiter before hitting the device;
// pacing and burst size are configured via the [Config] fields
// PauseBetweenRequests and BurstRequestCount. HTTP-level failures surface as
// the typed [ResponseError], and invalid configuration returns the exported
// ErrMissBaseURL or ErrMissUserOrPass sentinels.
//
// The upstream API is documented at
// https://help.mikrotik.com/docs/display/ROS/REST+API. API_COVERAGE.md in the
// repository tracks which endpoints are implemented and which are planned.
package routeros
