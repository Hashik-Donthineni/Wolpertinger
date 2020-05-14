package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
	AuthToken string `json:"auth_token"`
}

// ServerResponse is the response to a ClientRequest.  It maps a bridge's ID to
// a Bridge struct.
type ServerResponse map[string]*Bridge

// isRequestAuthenticated returns 'true' if we have the authentication token in
// the client request on record.
func isRequestAuthenticated(req *ClientRequest) bool {
	for _, t := range config.ApiTokens {
		if req.AuthToken == t.Token {
			return true
		}
	}
	return false
}

// IndexHandler handles requests for the service's index page.  We respond with
// a static string to make it easy for monitoring tools to check if
// wolpertinger is still alive and well.
func IndexHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintln(w, "Beware the Wolpertinger.")
}

// ProbeHandler deals with clients (e.g., an OONI probe) requesting a bridge to
// probe.
func ProbeHandler(w http.ResponseWriter, r *http.Request) {
	var req ClientRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("Received invalid JSON request.")
		http.Error(w, "invalid json request", http.StatusBadRequest)
		return
	}

	if !isRequestAuthenticated(&req) {
		log.Printf("Received unauthenticated request.")
		http.Error(w, "invalid authentication token", http.StatusUnauthorized)
		return
	}

	bridges, err := GetBridges(&req)
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
	fmt.Fprintln(w, string(json))
}
