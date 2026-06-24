# RouterOS API Coverage

This document tracks the implementation status of RouterOS API endpoints in this Go client library.

## Implemented Endpoints

### IP Section

- [x] `/ip/firewall/address-list` - Manage IP firewall address lists
- [x] `/ip/address` - Manage IP addresses

### Tool Section

- [x] `/tool/ping` - Ping hosts

## Planned Endpoints

### IP Section

- [ ] `/ip/route` - Manage IP routes
- [ ] `/ip/dns` - DNS configuration
- [ ] `/ip/dhcp-server` - DHCP server configuration
- [ ] `/ip/dhcp-client` - DHCP client configuration
- [ ] `/ip/hotspot` - Hotspot configuration
- [ ] `/ip/pool` - IP pools
- [ ] `/ip/service` - IP services (telnet, ftp, www, ssh, etc.)
- [ ] `/ip/accounting` - IP accounting
- [ ] `/ip/cloud` - Cloud settings
- [ ] `/ip/tunnel` - IP tunnels
- [ ] `/ip/upnp` - UPnP settings
- [ ] `/ip/traffic-flow` - Traffic flow (NetFlow) settings

### Interface Section

- [ ] `/interface` - Network interfaces
- [ ] `/interface/bridge` - Bridge interfaces
- [ ] `/interface/ethernet` - Ethernet interfaces
- [ ] `/interface/wireless` - Wireless interfaces
- [ ] `/interface/vlan` - VLAN interfaces
- [ ] `/interface/pppoe-client` - PPPoE client interfaces
- [ ] `/interface/l2tp-server` - L2TP server
- [ ] `/interface/ovpn-server` - OpenVPN server

### System Section

- [ ] `/system/identity` - System identity
- [ ] `/system/clock` - System clock
- [ ] `/system/resource` - System resources
- [ ] `/system/scheduler` - Scheduler
- [ ] `/system/script` - Scripts
- [ ] `/system/backup` - System backups
- [ ] `/system/logging` - System logging
- [ ] `/system/ntp` - NTP settings
- [ ] `/system/package` - System packages
- [ ] `/system/routerboard` - RouterBOARD information

### User Management

- [ ] `/user` - User management
- [ ] `/user/group` - User groups
- [ ] `/user/aaa` - Authentication, Authorization, Accounting

### Firewall Section

- [ ] `/ip/firewall/filter` - Firewall filters
- [ ] `/ip/firewall/nat` - NAT rules
- [ ] `/ip/firewall/mangle` - Packet mangling
- [ ] `/ip/firewall/raw` - Raw firewall rules
- [ ] `/ip/firewall/connection` - Firewall connections
- [ ] `/ip/firewall/layer7-protocol` - Layer 7 protocols
- [ ] `/ip/firewall/service-port` - Service ports

### Queue Section

- [ ] `/queue/simple` - Simple queues
- [ ] `/queue/tree` - Queue trees
- [ ] `/queue/type` - Queue types

### Routing Section

- [ ] `/routing/bgp` - BGP configuration
- [ ] `/routing/ospf` - OSPF configuration
- [ ] `/routing/rip` - RIP configuration
- [ ] `/routing/table` - Routing tables
- [ ] `/routing/filter` - Routing filters
- [ ] `/routing/rule` - Routing rules

### Tool Section

- [ ] `/tool/traceroute` - Traceroute tool
- [ ] `/tool/bandwidth-test` - Bandwidth testing
- [ ] `/tool/sniffer` - Packet sniffer
- [ ] `/tool/torch` - Torch (real-time traffic monitor)
- [ ] `/tool/mac-server` - MAC server
- [ ] `/tool/graphing` - Graphing

### PPP Section

- [ ] `/ppp/profile` - PPP profiles
- [ ] `/ppp/secret` - PPP secrets
- [ ] `/ppp/active` - Active PPP connections

### Wireless Section

- [ ] `/interface/wireless/security-profiles` - Wireless security profiles
- [ ] `/interface/wireless/access-list` - Wireless access list
- [ ] `/interface/wireless/connect-list` - Wireless connect list
- [ ] `/interface/wireless/default-ap-tx-chain` - Default AP TX chains

### Certificate Section

- [ ] `/certificate` - Certificates
- [ ] `/certificate/scep-server` - SCEP server

### SNMP Section

- [ ] `/snmp` - SNMP settings
- [ ] `/snmp/community` - SNMP communities

### MPLS Section

- [ ] `/mpls` - MPLS settings
- [ ] `/mpls/ldp` - LDP settings
- [ ] `/mpls/static` - Static MPLS labels

### IPv6 Section

- [ ] `/ipv6/address` - IPv6 addresses
- [ ] `/ipv6/route` - IPv6 routes
- [ ] `/ipv6/dhcp-server` - IPv6 DHCP server
- [ ] `/ipv6/dhcp-client` - IPv6 DHCP client
- [ ] `/ipv6/firewall/filter` - IPv6 firewall filters
- [ ] `/ipv6/firewall/nat` - IPv6 NAT rules
- [ ] `/ipv6/firewall/mangle` - IPv6 packet mangling
- [ ] `/ipv6/nd` - IPv6 Neighbor Discovery
- [ ] `/ipv6/dhcp-client/option` - IPv6 DHCP client options

### LTE Section

- [ ] `/interface/lte` - LTE interfaces
- [ ] `/interface/lte/apn` - LTE APNs

### Container Section

- [ ] `/container` - Containers
- [ ] `/container/config` - Container configurations
- [ ] `/container/env` - Container environment variables

### Other Sections

- [ ] `/caps-man` - CAPsMAN (Centralized WLAN management)
- [ ] `/radius` - RADIUS configuration
- [ ] `/radius/incoming` - Incoming RADIUS requests
- [ ] `/email` - Email settings
- [ ] `/sms` - SMS settings
- [ ] `/lcd` - LCD configuration
- [ ] `/led` - LED settings
- [ ] `/watchdog` - Watchdog settings
- [ ] `/idp` - Intrusion Detection Prevention
- [ ] `/romon` - RoMON (Router Monitoring)
- [ ] `/netwatch` - Netwatch (network monitoring)

## Implementation Priority

Based on common use cases and functionality, the following endpoints should be implemented next:

1. `/ip/route` - Essential for network configuration
2. `/interface` - Basic interface management
3. `/system/identity` - Basic system configuration
4. `/user` - User management
5. `/ip/firewall/filter` - Core firewall functionality
6. `/ip/dhcp-server` - DHCP services
7. `/interface/wireless` - Wireless configuration
8. `/queue/simple` - Traffic shaping
9. `/routing/ospf` - Dynamic routing
10. `/tool/traceroute` - Network diagnostics

## Contributing

To contribute new endpoint implementations:

1. Follow the existing code patterns in the repository
2. Create structs for request/response types in the `types` package
3. Implement the service methods in a new file (e.g., `ip_route.go`)
4. Add corresponding mock server endpoints in `pkg/mockserver`
5. Create comprehensive tests with mock server implementations
6. Update this document to reflect the new implementation status
7. Update README.md with usage examples if needed

## References

- [Official RouterOS REST API Documentation](https://help.mikrotik.com/docs/display/ROS/REST+API)
- [RouterOS Manual](https://wiki.mikrotik.com/wiki/Manual:Contents)
