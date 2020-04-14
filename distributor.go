package main

import (
	"net"
)

// GetBridges queries our backend to find and return bridges that we want
// tested by censorship measurement platforms like OONI.
func GetBridges(req *ClientRequest) ([]Bridge, error) {

	// FIXME: This is placeholder code, just so we have something to work with.
	v4, _ := net.ResolveIPAddr("", "1.2.3.4")
	v6, _ := net.ResolveIPAddr("", "2001:4860:4860::8888")

	obfs4Args := map[string]string{
		"cert":     "VBYOXYf+SbRu2dCHJkLuL9y7YX4IWhucHGg3ES+l/KKxe3KL+zhCHr5hRqgSE6w80bZvCA",
		"iat-mode": "0",
	}
	b1 := Bridge{
		BridgeTypeVanilla,
		ProtoTypeTCP,
		IPAddr{net.IPAddr{v4.IP, v4.Zone}},
		443,
		"1234567890ABCDEF1234567890ABCDEF12345678",
		nil}
	b2 := Bridge{
		BridgeTypeObfs4,
		ProtoTypeTCP,
		IPAddr{net.IPAddr{v6.IP, v6.Zone}},
		80,
		"1234567890ABCDEF1234567890ABCDEF12345678",
		obfs4Args}

	return []Bridge{b1, b2}, nil
}
