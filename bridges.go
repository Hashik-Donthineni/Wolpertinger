package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
)

const (
	BridgeTypeVanilla = "vanilla"
	BridgeTypeObfs4   = "obfs4"
)

const (
	ProtoTypeTCP = "tcp"
	ProtoTypeUDP = "udp"
)

// IPAddr embeds net.IPAddr.  The only difference to net.IPAddr is that we
// implement a MarshalJSON method that allows for convenient marshalling of IP
// addresses.
type IPAddr struct {
	net.IPAddr
}

func (a IPAddr) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

// Bridge represents a Tor bridge.  It consists of a Type (e.g., "vanilla" or
// "obfs4"), a Protocol (e.g., "udp" or "tcp"), an Address, Port, Fingerprint,
// and Arguments (optional).
type Bridge struct {
	Type        string            `json:"type"`
	Protocol    string            `json:"protocol"`
	Address     IPAddr            `json:"address"`
	Port        uint16            `json:"port"`
	Fingerprint string            `json:"fingerprint"`
	Arguments   map[string]string `json:"arguments,omitempty"`
}

// GetID returns a unique ID that we derive from a bridge's three-tuple (i.e.,
// its IP address, port, and protocol).  We cannot take the bridge's
// fingerprint because all its PTs share the fingerprint.  We derive the unique
// ID by doing a HMAC (keyed with a master secret from our config file) over
// the bridge's three-tuple.
func (b *Bridge) GetID() string {

	threeTuple := fmt.Sprintf("%s-%d-%d", b.Address.String(), b.Port, b.Protocol)

	h := hmac.New(sha256.New, []byte(config.MasterKey))
	h.Write([]byte(threeTuple))

	return hex.EncodeToString(h.Sum(nil))
}
