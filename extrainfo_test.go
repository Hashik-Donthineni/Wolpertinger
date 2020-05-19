package main

import (
	"bytes"
	"testing"
)

func TestPopulateTransportInfo(t *testing.T) {

	var err error
	var transport = NewTransport()

	if err = populateTransportInfo("", transport); err == nil {
		t.Errorf("Failed to fail when given empty transport line.")
	}

	if err = populateTransportInfo("transport", transport); err == nil {
		t.Errorf("Failed to fail when given invalid transport line.")
	}

	if err = populateTransportInfo("transport foo 1.2.3.4:1234", transport); err != nil {
		t.Errorf("Failed to parse transport line.")
	}
	if transport.Type != "foo" {
		t.Errorf("Failed to parse transport type.")
	}
	if transport.Port != 1234 {
		t.Errorf("Failed to parse transport port.")
	}

	if err = populateTransportInfo("transport bar 1.2.3.4:1234 a=b,foo=bar", transport); err != nil {
		t.Errorf("Failed to parse transport line.")
	}
	value, _ := transport.Parameters["a"]
	if value[0] != "b" {
		t.Errorf("Failed to parse transport arguments.")
	}
	value, _ = transport.Parameters["foo"]
	if value[0] != "bar" {
		t.Errorf("Failed to parse transport arguments.")
	}
}

func TestParseExtrainfoDoc(t *testing.T) {

	var bridges *Bridges
	var err error
	buf := bytes.NewBufferString(`extra-info foo A0EC5B0FC51A5CD800B9D1D16D325636B5755BCE
this line doesn't matter
and neither does this one
extra-info bar 51502DF3D176CC10C52CC65694205BBA185E0982
transport obfs4 1.2.3.4:1234 key=value,1=2
transport obfs5 1.2.3.4:4321 foo=bar
`)

	bridges, err = ParseExtrainfoDoc(buf)
	if err != nil {
		t.Errorf("Failed to parse mock extra-info descriptors.")
	}

	if len(bridges.Bridges) != 2 {
		t.Errorf("Parsed incorrect number of bridges.")
	}
	bridge, ok := bridges.Bridges["51502DF3D176CC10C52CC65694205BBA185E0982"]
	if !ok {
		t.Errorf("Fingerprint doesn't exist in bridges map.")
	}
	if len(bridge.Transports) != 2 {
		t.Errorf("Bridge has incorrect number of transports.")
	}
	if bridge.Transports[0].Type != BridgeTypeObfs4 {
		t.Errorf("Didn't parse obfs4 transport line.")
	}
	if bridge.Transports[0].Port != 1234 {
		t.Errorf("Couldn't parse obfs4 port.")
	}
	if bridge.Transports[0].Fingerprint != "51502DF3D176CC10C52CC65694205BBA185E0982" {
		t.Errorf("Couldn't parse obfs4 fingerprint.")
	}
	value, _ := bridge.Transports[0].Parameters["key"]
	if value[0] != "value" {
		t.Errorf("Couldn't parse first obfs4 key=value pair.")
	}
	value, _ = bridge.Transports[0].Parameters["1"]
	if value[0] != "2" {
		t.Errorf("Couldn't parse second obfs4 key=value pair.")
	}
}
