package main

import (
	"database/sql"
	"fmt"
	"net"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

const (
	LastSeenLayout = "2006-01-02 15:04"
)

func LoadDatabase(db *sql.DB) (*Bridges, error) {

	// Figure out the latest 'last_seen' value, which allows us to select only
	// bridges that are currently online.
	var latest string
	rows, err := db.Query("SELECT last_seen FROM Bridges ORDER BY last_seen DESC LIMIT 1;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&latest); err != nil {
			return nil, err
		}
	}

	// Now select the actual bridges.
	query := fmt.Sprintf("SELECT * FROM Bridges WHERE last_seen = '%s' AND or_port IS NOT NULL;", latest)
	rows, err = db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bridges = NewBridges()
	var b *Bridge
	var id, fingerprint, address, port, distributor, first_seen, last_seen string

	for rows.Next() {
		b = NewBridge()
		err = rows.Scan(&id, &fingerprint, &address, &port, &distributor, &first_seen, &last_seen)
		if err != nil {
			return nil, err
		}
		b.Fingerprint = fingerprint
		b.Distributor = distributor

		a, err := net.ResolveIPAddr("", address)
		if err != nil {
			return nil, err
		}
		b.Address = IPAddr{net.IPAddr{a.IP, a.Zone}}
		p, err := strconv.Atoi(port)
		if err != nil {
			return nil, err
		}
		b.Port = uint16(p)

		bridges.Bridges[b.Fingerprint] = b
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return bridges, nil
}

func InsertBlockedBridge(db *sql.DB) error {

	// TODO: Take as input some sort of bridge object and write it to the given
	// SQLite database.
	return nil
}
