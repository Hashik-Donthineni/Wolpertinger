package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	//"os"
	"strconv"
	"strings"
)

const (
	MinTransportWords = 3
	TransportPrefix   = "transport"
	ExtraInfoPrefix   = "extra-info"
)

// populateTransportInfo parses the given transport line of the format:
//   "transport" transportname address:port [arglist] NL
// ...and writes it to the given transport object.  See the specification for
// more details on what transport lines look like:
// <https://gitweb.torproject.org/torspec.git/tree/dir-spec.txt?id=2b31c63891a63cc2cad0f0710a45989071b84114#n1234>
func populateTransportInfo(transport string, t *Transport) error {

	if !strings.HasPrefix(transport, TransportPrefix) {
		return errors.New("no 'transport' prefix")
	}

	words := strings.Split(transport, " ")
	if len(words) < MinTransportWords {
		return errors.New("not enough arguments in 'transport' line")
	}
	t.Type = words[1]

	host, port, err := net.SplitHostPort(words[2])
	if err != nil {
		return err
	}
	addr, err := net.ResolveIPAddr("", host)
	if err != nil {
		return err
	}
	t.Address = IPAddr{net.IPAddr{addr.IP, addr.Zone}}
	p, err := strconv.Atoi(port)
	if err != nil {
		return err
	}
	t.Port = uint16(p)

	// We may be dealing with one or more key=value pairs.
	if len(words) > MinTransportWords {
		args := strings.Split(words[3], ",")
		for _, arg := range args {
			kv := strings.Split(arg, "=")
			if len(kv) != 2 {
				return fmt.Errorf("key:value pair in %q not separated by a '='", words[3])
			}
			t.Arguments[kv[0]] = kv[1]
		}
	}

	return nil
}

// ParseExtrainfoDoc parses the given extra-info document and returns the
// content as a Bridges object.  Note that the extra-info document format is as
// it's produced by the bridge authority.
func ParseExtrainfoDoc(r io.Reader) (*Bridges, error) {

	var bridges = NewBridges()
	var b *Bridge

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		// We're dealing with a new extra-info block, i.e., a new bridge.
		if strings.HasPrefix(line, ExtraInfoPrefix) {
			b = NewBridge()
			words := strings.Split(line, " ")
			if len(words) != 3 {
				return nil, errors.New("incorrect number of words in 'extra-info' line")
			}
			b.Fingerprint = words[2]
			bridges.Bridges[b.Fingerprint] = b
		}
		// We're dealing with a bridge's transport protocols.  There may be
		// several.
		if strings.HasPrefix(line, TransportPrefix) {
			t := NewTransport()
			t.Fingerprint = b.Fingerprint
			err := populateTransportInfo(line, t)
			if err != nil {
				return nil, err
			}
			b.AddTransport(t)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return bridges, nil
}
