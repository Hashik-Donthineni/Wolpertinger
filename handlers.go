package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
)

const (
	ProbeTypeOONI = "ooni"
)

// ClientRequest represents a request to probe a bridge, e.g., an OONI probe
// asking to get a bridge to test.
type ClientRequest struct {
	Id        string `json:"id"`
	ProbeType string `json:"type"`
	Location  string `json:"country_code"`
}

// ServerResponse is the response to a ClientRequest.  It maps a bridge's ID to
// a Bridge struct.
type ServerResponse map[string]*Bridge

// authenticateRequest attempts to authenticate the given HTTP request.  If
// this fails, it returns an error and an HTTP status code that should be
// returned to the client.
func authenticateRequest(r *http.Request) (error, int) {

	// First, we get the bearer token from the 'Authorization' HTTP header.
	tokenLine := r.Header.Get("Authorization")
	if tokenLine == "" {
		return errors.New("request has not 'Authorization' HTTP header"), http.StatusBadRequest
	}
	if !strings.HasPrefix(tokenLine, "Bearer ") {
		return errors.New("authorization header contains no bearer token"), http.StatusBadRequest
	}
	fields := strings.Split(tokenLine, " ")
	token := fields[1]

	for _, t := range config.ApiTokens {
		if token == t.Token {
			return nil, 0
		}
	}
	return errors.New("invalid authentication token"), http.StatusUnauthorized
}

// IndexHandler handles requests for the service's index page.  We respond with
// a static string to make it easy for monitoring tools to check if
// wolpertinger is still alive and well.
func IndexHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintln(w, "Beware the Wolpertinger.")
}

// extractClientRequest attempts to extract a ClientRequest object from the
// given HTTP request.
func extractClientRequest(r *http.Request) (*ClientRequest, error) {

	// Now get our request fields, which are in the GET request URL.
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	id, ok := r.Form["id"]
	if !ok {
		return nil, errors.New("key 'id' not found in request")
	}

	reqType, ok := r.Form["type"]
	if !ok {
		return nil, errors.New("key 'type' not found in request")
	} else if len(reqType) != 1 {
		return nil, errors.New("need exactly one 'type' key")
	}

	countryCode, ok := r.Form["country_code"]
	if !ok {
		return nil, errors.New("key 'country_code' not found in request")
	} else if len(countryCode) != 1 {
		return nil, errors.New("need exactly one 'country_code' key")
	}

	return &ClientRequest{id[0], reqType[0], countryCode[0]}, nil
}

// BridgesHandler deals with clients (e.g., an OONI probe) requesting a bridge
// to probe.
func BridgesHandler(w http.ResponseWriter, r *http.Request) {

	if err, statusCode := authenticateRequest(r); err != nil {
		http.Error(w, err.Error(), statusCode)
		return
	}

	req, err := extractClientRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	bridges, err := GetBridges(req)
	if err != nil {
		log.Printf("Error getting bridges: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := ServerResponse{}
	for _, bridge := range bridges.Bridges {
		resp[bridge.GetID()] = bridge
	}

	json, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintln(w, string(json))
}
