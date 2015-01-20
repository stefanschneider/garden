package garden

import (
	"fmt"
	"net"
)

func lastIP(n net.IPNet) net.IP {
	mask := n.Mask
	ip := n.IP
	lastip := make(net.IP, length(ip))
	for i, m := range mask {
		lastip[i] = (^mask[i]) | ip[i]
	}
	return lastip
}

func (pr PortRange) String() string {
	if pr.Start == 0 && pr.End == 0 {
		return ""
	}
	return fmt.Sprintf("%d:%d", pr.Start, pr.End)
}

const (
	icmpAllTypes int32 = -1
	icmpAllCodes int32 = -1
)

func ICMPType(t int32) *iCMPType {
	p := iCMPType(t)
	return &p
}

func ICMPCode(c int32) *iCMPCode {
	p := iCMPCode(c)
	return &p
}

type iCMPType int32
type iCMPCode int32

func (t *iCMPType) icmpType() int32 {
	if t == nil {
		return icmpAllTypes
	}
	return int32(*t)
}

func (c *iCMPCode) icmpCode() int32 {
	if c == nil {
		return icmpAllCodes
	}
	return int32(*c)
}

func (r NetOutRule) Rule() NetOutRule {
	return r
}

func (r AllRule) Rule() NetOutRule {
	return NetOutRule{
		Network:   r.Network,
		Port:      0,
		PortRange: PortRange{},
		Protocol:  ProtocolAll,
		IcmpType:  icmpAllTypes,
		IcmpCode:  icmpAllCodes,
		Log:       r.Log,
	}
}

func (r UDPRule) Rule() NetOutRule {
	return NetOutRule{
		Network:   r.Network,
		Port:      r.Port,
		PortRange: r.PortRange,
		Protocol:  ProtocolUDP,
		IcmpType:  icmpAllTypes,
		IcmpCode:  icmpAllCodes,
		Log:       false,
	}
}

func (r ICMPRule) Rule() NetOutRule {
	return NetOutRule{
		Network:   r.Network,
		Port:      0,
		PortRange: PortRange{},
		Protocol:  ProtocolICMP,
		IcmpType:  r.Type.icmpType(),
		IcmpCode:  r.Code.icmpCode(),
		Log:       false,
	}
}

func (r TCPRule) Rule() NetOutRule {
	return NetOutRule{
		Network:   r.Network,
		Port:      r.Port,
		PortRange: r.PortRange,
		Protocol:  ProtocolTCP,
		IcmpType:  icmpAllTypes,
		IcmpCode:  icmpAllCodes,
		Log:       r.Log,
	}
}
