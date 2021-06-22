package ip_utils

import (
	"net"
)

/*
Generate list of IPv4 from network
*/
func LookupHost(ipnet *net.IPNet) []net.IP {
	ip := ipnet.IP
	startIp := int(ip[0])<<24 + int(ip[1])<<16 + int(ip[2])<<8 + int(ip[3])

	ipList := []net.IP{}
	for i := startIp; i < 0xffffffff; i++ {
		oc4 := (i >> 24) & 0xff
		oc3 := (i >> 16) & 0xff
		oc2 := (i >> 8) & 0xff
		oc1 := i & 0xff

		ip4 := net.IPv4(byte(oc4), byte(oc3), byte(oc2), byte(oc1))

		if ipnet.Contains(ip4) {
			ipList = append(ipList, ip4)
		} else {
			break
		}
	}
	return ipList
}
