package main

import (
	"fmt"
	"net/http"
	"testing"
)

func TestExtractClientRequest(t *testing.T) {

	var baseUrl = "https://bridges.torproject.org/wolpertinger/bridges"

	req, _ := http.NewRequest("GET", baseUrl, nil)
	_, err := extractClientRequest(req)
	if err == nil {
		t.Error("accepted request with no arguments")
	}

	req, _ = http.NewRequest("GET", fmt.Sprintf("%s?id=1234&type=foo", baseUrl), nil)
	_, err = extractClientRequest(req)
	if err == nil {
		t.Error("accepted request with missing arguments")
	}

	req, _ = http.NewRequest("GET", fmt.Sprintf("%s?id=1234&type=foo&type=bar", baseUrl), nil)
	_, err = extractClientRequest(req)
	if err == nil {
		t.Error("accepted request with duplicate argument")
	}

	req, _ = http.NewRequest("GET", fmt.Sprintf("%s?id=1234&type=foo&country_code=ru", baseUrl), nil)
	ret, err := extractClientRequest(req)
	if err != nil {
		t.Errorf("failed to accept valid arguments: %s", err.Error())
	}
	if ret.Id != "1234" {
		t.Error("failed to parse client ID")
	}
	if ret.ProbeType != "foo" {
		t.Error("failed to parse probe type")
	}
	if ret.Location != "ru" {
		t.Error("failed to parse country code")
	}

	req, _ = http.NewRequest("GET", fmt.Sprintf("%s?id=&type=foo&country_code=ru", baseUrl), nil)
	_, err = extractClientRequest(req)
	if err != nil {
		t.Errorf("failed to accept empty id argument: %s", err.Error())
	}
}

func TestRequestAuthorization(t *testing.T) {

	var url = "https://bridges.torproject.org/wolpertinger/bridges"
	var apiToken = "KEWDlzJ7JLCBZ2dJ6pXa4P04aq0rbi1weJXGBAP0H/o="

	config = ConfigFile{
		"bogus master key",
		[]ApiToken{ApiToken{"foo", apiToken}},
		"bogus sqlite file",
		"bogus extrainfo file",
	}

	req, _ := http.NewRequest("GET", url, nil)
	err, status := authenticateRequest(req)
	if err == nil || status != http.StatusBadRequest {
		t.Error("failed to reject invalid request")
	}

	req.Header.Set("Authorization", "Foo")
	err, status = authenticateRequest(req)
	if err == nil || status != http.StatusBadRequest {
		t.Error("failed to reject invalid request")
	}

	req.Header.Set("Authorization", "Bearer Foo")
	err, status = authenticateRequest(req)
	if err == nil || status != http.StatusUnauthorized {
		t.Error("failed to reject invalid authentication token")
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiToken))
	err, status = authenticateRequest(req)
	if err != nil {
		t.Error("failed to accept valid request")
	}
}
