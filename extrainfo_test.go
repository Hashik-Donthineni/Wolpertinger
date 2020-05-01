package main

import (
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
	value, _ := transport.Arguments["a"]
	if value != "b" {
		t.Errorf("Failed to parse transport arguments.")
	}
	value, _ = transport.Arguments["foo"]
	if value != "bar" {
		t.Errorf("Failed to parse transport arguments.")
	}
}
