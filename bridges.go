package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"reflect"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	BridgeTypeVanilla = "vanilla"
	BridgeTypeObfs4   = "obfs4"

	ProtoTypeTCP = "tcp"
	ProtoTypeUDP = "udp"

	DistributorMoat        = "moat"
	DistributorHttps       = "https"
	DistributorEmail       = "email"
	DistributorUnallocated = "unallocated"

	BridgeReloadInterval = time.Hour
)

// bridges holds all of our bridges.
var bridges Bridges

// IPAddr embeds net.IPAddr.  The only difference to net.IPAddr is that we
// implement a MarshalJSON method that allows for convenient marshalling of IP
// addresses.
type IPAddr struct {
	net.IPAddr
}

func (a IPAddr) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

// Bridges represents a set of Bridge objects.
type Bridges struct {
	m       sync.Mutex
	Bridges map[string]*Bridge
}

// Update allows us to update our set of bridges.
func (old *Bridges) Update(new *Bridges) {
	old.m.Lock()
	old.Bridges = new.Bridges
	old.m.Unlock()
}

// Add adds the given bridge to the set of bridges.
func (bs *Bridges) Add(b *Bridge) {
	bs.Bridges[b.Fingerprint] = b
}

// NewBridges allocates and returns a new Bridges object.
func NewBridges() *Bridges {
	b := &Bridges{}
	b.Bridges = make(map[string]*Bridge)
	return b
}

// Location represents a location in which a bridge is blocked.  This is either
// a two-letter country code or an AS number.
type Location struct {
	Country string // An ISO 3166-1 alpha-2 country code.
	ASN     int    // An autonomous system number.
}

// Bridge represents a Tor bridge.
type Bridge struct {
	Type        string       `json:"type"`
	Protocol    string       `json:"protocol"`
	Address     IPAddr       `json:"address"`
	Port        uint16       `json:"port"`
	Fingerprint string       `json:"fingerprint"`
	Distributor string       `json:"-"`
	FirstSeen   time.Time    `json:"-"`
	LastSeen    time.Time    `json:"-"`
	BlockedIn   []*Location  `json:"-"`
	Transports  []*Transport `json:"-"`
}

// String returns a string representation of the bridge.
func (b *Bridge) String() string {
	return fmt.Sprintf("%s (%s)\n\t%s:%d",
		b.Fingerprint, b.Distributor, b.Address.String(), b.Port)
}

// NewBridge allocates and returns a new Bridge object.
func NewBridge() *Bridge {
	b := &Bridge{}
	// A bridge (without pluggable transports) is always running vanilla Tor
	// over TCP.
	b.Protocol = ProtoTypeTCP
	b.Type = BridgeTypeVanilla
	return b
}

// AddTransport adds the given transport to the bridge.
func (b *Bridge) AddTransport(t1 *Transport) {
	for _, t2 := range b.Transports {
		if reflect.DeepEqual(t1, t2) {
			// We already have this transport on record.
			return
		}
	}
	b.Transports = append(b.Transports, t1)
}

// GetID returns a unique ID that we derive from a bridge's three-tuple (i.e.,
// its IP address, port, and protocol).  We derive the unique ID by doing a
// HMAC (keyed with a master secret from our config file) over the bridge's
// three-tuple.
func (b *Bridge) GetID() string {

	threeTuple := fmt.Sprintf("%s-%d-%s", b.Address.String(), b.Port, ProtoTypeTCP)
	return Hmac([]byte(threeTuple))
}

func (bs *Bridges) ReloadBridges(done chan bool) {

	ticker := time.NewTicker(BridgeReloadInterval)
	defer ticker.Stop()

	sentDone := false
	for ; true; <-ticker.C {
		db, err := sql.Open("sqlite3", config.SqliteFile)
		if err != nil {
			log.Printf("Failed to open SQLite database: %s", err)
			continue
		}
		defer db.Close()
		sql, err := LoadDatabase(db)
		if err != nil {
			log.Printf("Failed to read bridges from SQLite database: %s", err)
			continue
		}

		file, err := os.Open(config.ExtrainfoFile)
		if err != nil {
			log.Printf("Failed to open extrainfo file: %s", err)
			continue
		}
		defer file.Close()
		extra, err := ParseExtrainfoDoc(file)
		if err != nil {
			log.Printf("Failed to read bridges from extrainfo file: %s", err)
			continue
		}

		for f, b1 := range sql.Bridges {
			// Do we have any transports for this bridge?
			if b2, ok := extra.Bridges[f]; ok {
				b1.Transports = b2.Transports
			}
		}

		log.Printf("Successfully loaded %d bridges.", len(sql.Bridges))
		bs.Update(sql)

		// Once, after our very first run, we signal to the caller that we're
		// done.  The caller can the proceed to start the REST API.
		if !sentDone {
			done <- true
			sentDone = true
		}
	}
}
