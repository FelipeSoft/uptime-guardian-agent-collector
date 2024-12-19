package network

import (
	"fmt"
	"net"
)

func GetLocalIPv4InSubnet(subnet string) ([]net.IP, error) {
	_, targetNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, fmt.Errorf("invalid subnet: %v", err)
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, fmt.Errorf("error getting network interface addresses: %v", err)
	}

	var ips []net.IP
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil && targetNet.Contains(ipnet.IP) {
				ips = append(ips, ipnet.IP)
			}
		}
	}

	if len(ips) == 0 {
		return nil, fmt.Errorf("no valid local IPv4 addresses found in subnet %s", subnet)
	}

	return ips, nil
}
