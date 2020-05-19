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

	var apiToken = "KEWDlzJ7JLCBZ2dJ6pXa4P04aq0rbi1weJXGBAP0H/o="
	config = ConfigFile{
		"bogus master key",
		[]ApiToken{ApiToken{"foo", apiToken}},
		"bogus sqlite file",
		"bogus extrainfo file",
	}

	req, _ = http.NewRequest("GET", fmt.Sprintf("%s?id=1234&type=foo&country_code=ru", baseUrl), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiToken))
	ret, err := extractClientRequest(req)
	if err != nil {
		t.Errorf("failed to accept valid arguments: %s", err.Error())
	}
	if ret.Id != "1234" {
		t.Errorf("failed to parse client ID")
	}
	if ret.ProbeType != "foo" {
		t.Errorf("failed to parse probe type")
	}
	if ret.Location != "ru" {
		t.Errorf("failed to parse country code")
	}
	if ret.AuthToken != apiToken {
		t.Errorf("failed to parse bearer token")
	}

	req, _ = http.NewRequest("GET", fmt.Sprintf("%s?id=&type=foo&country_code=ru", baseUrl), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiToken))
	_, err = extractClientRequest(req)
	if err != nil {
		t.Errorf("failed to accept empty id argument: %s", err.Error())
	}
}
