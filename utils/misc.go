package utils

import "net"

// ParseCIDRAsIPNet parse given CIDR into an IPNet but with its IP set to the address not the network.
// For example, for 192.168.1.1/24 net.ParseCIDR returns 192.168.1.0/24 but this function gives 192.168.1.0/24
func ParseCIDRAsIPNet(v string) (ret *net.IPNet, err error) {
	var ip net.IP
	if ip, ret, err = net.ParseCIDR(v); err != nil {
		return
	} else {
		ret.IP = ip
		return
	}
}
