package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
)

var config ConfigFile

type ConfigFile struct {
	MasterKey string     `json:"master_key"`
	ApiTokens []ApiToken `json:"api_tokens"`
}

type ApiToken struct {
	Organisation string `json:"organisation"`
	Token        string `json:"token"`
}

// loadConfigFile loads our JSON-encoded configuration file from disk.
func loadConfigFile(filename string) error {

	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(content, &config); err != nil {
		return err
	}

	return nil
}

func main() {

	var addr string
	var certFilename, keyFilename string
	var configFilename string

	flag.StringVar(&addr, "addr", ":7000", "Address to listen on.")
	flag.StringVar(&certFilename, "cert", "", "TLS certificate file.")
	flag.StringVar(&keyFilename, "key", "", "TLS private key file.")
	flag.StringVar(&configFilename, "config", "", "Configuration file.")
	flag.Parse()

	if configFilename == "" {
		log.Fatal("No configuration file given.")
	} else {
		if err := loadConfigFile(configFilename); err != nil {
			log.Fatalf("Failed to load config file: %s", err)
		}
	}

	mux := http.NewServeMux()
	mux.Handle("/api/fetch", http.HandlerFunc(ProbeHandler))

	log.Printf("Starting service on %s.", addr)
	if certFilename != "" && keyFilename != "" {
		log.Fatal(http.ListenAndServeTLS(addr, certFilename, keyFilename, mux))
	} else {
		log.Fatal(http.ListenAndServe(addr, mux))
	}
}
