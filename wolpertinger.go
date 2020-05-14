package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const (
	AuthTokenSize = 32
)

var config ConfigFile

type ConfigFile struct {
	MasterKey     string     `json:"master_key"`
	ApiTokens     []ApiToken `json:"api_tokens"`
	SqliteFile    string     `json:"sqlite_file"`
	ExtrainfoFile string     `json:"extrainfo_file"`
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

// genNewToken generates and returns a new authentication token.
func genNewToken() (string, error) {

	buf := make([]byte, AuthTokenSize)
	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf), nil
}

func main() {

	var addr string
	var certFilename, keyFilename string
	var configFilename string
	var logFilename string
	var newToken bool

	flag.StringVar(&addr, "addr", ":7000", "Address to listen on.")
	flag.StringVar(&certFilename, "cert", "", "TLS certificate file.")
	flag.StringVar(&keyFilename, "key", "", "TLS private key file.")
	flag.StringVar(&configFilename, "config", "", "Configuration file.")
	flag.StringVar(&logFilename, "log", "", "Log file.")
	flag.BoolVar(&newToken, "new-token", false, "Generate a new authentication token.")
	flag.Parse()

	if logFilename != "" {
		f, err := os.OpenFile(logFilename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			log.Fatalf("Error opening file %v", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	if newToken {
		token, err := genNewToken()
		if err != nil {
			log.Fatalf("Failed to generate new authentication token: %s")
		}
		fmt.Printf("Authentication token: %s\n", token)
		return
	}

	if configFilename == "" {
		log.Fatal("No configuration file given.")
	} else {
		if err := loadConfigFile(configFilename); err != nil {
			log.Fatalf("Failed to load config file: %s", err)
		}
	}

	// (Re-)load bridges periodically.  We wait for this function to finish its
	// first run before proceeding to start our web service.
	done := make(chan bool)
	defer close(done)
	go bridges.ReloadBridges(done)
	<-done

	mux := http.NewServeMux()
	mux.Handle("/fetch", http.HandlerFunc(ProbeHandler))
	mux.Handle("/", http.HandlerFunc(IndexHandler))

	log.Printf("Starting service on %s.", addr)
	if certFilename != "" && keyFilename != "" {
		log.Fatal(http.ListenAndServeTLS(addr, certFilename, keyFilename, mux))
	} else {
		log.Fatal(http.ListenAndServe(addr, mux))
	}
}
