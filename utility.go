package main

import (
	"encoding/binary"
	"net"
)

func IP2Uint32(ip net.IP) uint32 {
	return binary.BigEndian.Uint32(ip.To4())
}
