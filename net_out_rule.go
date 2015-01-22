package garden

import "net"

// Helper functions for constructing NetOutRule structure fields

// AllNetworks is a NetworkInterval helper function
func AllNetworks() *NetworkInterval {
	return nil
}

// AllIPv4Networks is a NetworkInterval helper function
func AllIPv4Networks() *NetworkInterval {
	return IPRange(net.ParseIP("0.0.0.0"), net.ParseIP("255.255.255.255"))
}

// IPRange is a NetworkInterval helper function
func IPRange(start, end net.IP) *NetworkInterval {
	return &NetworkInterval{
		Start: start,
		End:   end,
	}
}

// SingleIP is a NetworkInterval helper function
func SingleIP(ip net.IP) *NetworkInterval {
	return IPRange(ip, ip)
}

// IPNetNetwork is a NetworkInterval helper function
func IPNetNetwork(ipNet net.IPNet) *NetworkInterval {
	return IPRange(ipNet.IP, lastIP(ipNet))
}

// AllPorts is a PortInterval helper function
func AllPorts() *PortInterval {
	return nil
}

// SinglePort is a PortInterval helper function
func SinglePort(port uint16) *PortInterval {
	return PortRange(port, port)
}

// PortRange is a PortInterval helper function
func PortRange(start, end uint16) *PortInterval {
	return &PortInterval{
		Start: start,
		End:   end,
	}
}

// AllICMPs is a ICMPControl helper function
func AllICMPs() *ICMPControl {
	return nil
}

// AllICMPsOfType is a ICMPControl helper function
func AllICMPsOfType(iType uint8) *ICMPControl {
	return &ICMPControl{
		Type: iType,
		Code: nil,
	}
}

// ICMPTypeAndCode is a ICMPControl helper function
func ICMPTypeAndCode(iType, iCode uint8) *ICMPControl {
	pCode := iCode
	return &ICMPControl{
		Type: iType,
		Code: &pCode,
	}
}

// Last IP (broadcast) address in a network (net.IPNet)
func lastIP(n net.IPNet) net.IP {
	mask := n.Mask
	ip := n.IP
	lastip := make(net.IP, len(ip))
	// set bits zero in the mask to ones in ip
	for i, m := range mask {
		lastip[i] = (^m) | ip[i]
	}
	return lastip
}
