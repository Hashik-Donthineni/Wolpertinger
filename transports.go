package main

import (
	"fmt"
	"reflect"
	"strings"
)

// Transport represents a Tor bridge's pluggable transport.
type Transport struct {
	Type        string              `json:"type"`
	Protocol    string              `json:"protocol"`
	Address     IPAddr              `json:"address"`
	Port        uint16              `json:"port"`
	Fingerprint string              `json:"fingerprint"`
	Parameters  map[string][]string `json:"params,omitempty"`
	Bridge      *Bridge             `json:"-"`
	BlockedIn   []*Location         `json:"-"`
}

// NewTransport allocates and returns a new Transport object.
func NewTransport() *Transport {
	t := &Transport{}
	t.Parameters = make(map[string][]string)
	return t
}

// String returns a string representation of the transport.
func (t *Transport) String() string {

	var args []string
	for key, values := range t.Parameters {
		args = append(args, fmt.Sprintf("%s=%s", key, values[0]))
	}

	return fmt.Sprintf("%s %s:%d %s %s",
		t.Type, t.Address.String(), t.Port, t.Fingerprint, strings.Join(args, ","))
}

// Equals returns 'true' if the two given transports are identical, i.e., the
// values in their respective structs are identical.
func (t1 *Transport) Equals(t2 *Transport) bool {
	return reflect.DeepEqual(t1, t2)
}

// GetID returns a unique ID that we derive from a transport's three-tuple
// (i.e., its IP address, port, and protocol).  We derive the unique ID by
// doing a HMAC (keyed with a master secret from our config file) over the
// bridge's three-tuple.
func (t *Transport) GetID() string {

	threeTuple := fmt.Sprintf("%s-%d-%s", t.Address.String(), t.Port, t.Protocol)
	return Hmac([]byte(threeTuple))
}
