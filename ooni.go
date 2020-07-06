package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mholt/archiver"
)

const (
	S3BaseUrl              = "http://s3.amazonaws.com/ooni-data/canned"
	TorTestName            = "tor"
	OoniDataReloadInterval = time.Hour * 24
)

// OoniMeasurement represents a measurement of an OONI probe.  Here's the data
// format specification:
// <https://github.com/ooni/spec/blob/master/data-formats/df-000-base.md>
// Note that we are only interested in a subset of the fields, hence the
// incomplete structure below.
type OoniMeasurement struct {
	ProbeAsn string         `json:"probe_asn"`
	ProbeCc  string         `json:"probe_cc"`
	TestName string         `json:"test_name"`
	TestKeys TorMeasurement `json:"test_keys"`
}

// TorMeasurement represents a set of Tor measurements.  Here's the data
// format specification:
// <https://github.com/ooni/spec/blob/master/nettests/ts-023-tor.md#expected-output>
type TorMeasurement struct {
	DirPortTotal            int                  `json:"dir_port_total"`
	DirPortAccessible       int                  `json:"dir_port_accessible"`
	Obfs4Total              int                  `json:"obfs4_total"`
	Obfs4Accessible         int                  `json:"obfs4_accessible"`
	OrPortDirauthTotal      int                  `json:"or_port_dirauth_total"`
	OrPortDirauthAccessible int                  `json:"or_port_dirauth_accessible"`
	OrPortTotal             int                  `json:"or_port_total"`
	OrPortAccessible        int                  `json:"or_port_accessible"`
	Targets                 map[string]TorTarget `json:"targets"`
}

// TorTarget represents a Tor measurement target.  That can be a directory
// authority or a bridge.  Here's the data format specification:
// <https://github.com/ooni/spec/blob/master/nettests/ts-023-tor.md#expected-output>
type TorTarget struct {
	Failure        string             `json:"failure"`
	Summary        map[string]Failure `json:"summary"`
	TargetAddress  string             `json:"target_address"` // Expressed as 1.1.1.1:555, [::1]:555, or domain:555.
	TargetName     string             `json:"target_name"`
	TargetProtocol string             `json:"target_protocol"`
}

// Failure maps the "failure" key to one of the errors discussed in the
// following specification:
// <https://github.com/ooni/spec/blob/master/data-formats/df-007-errors.md>
type Failure map[string]string

type OoniBucket struct {
	FileSize  int    `json:"file_size"`
	TextSize  int    `json:"text_size"`
	FileSha1  string `json:"file_sha1"`
	FileCrc32 int    `json:"file_crc32"`
	FileName  string `json:"filename"`
}

func (f Failure) String() string {

	return f["failure"]
}

func (t TorTarget) String() string {

	return fmt.Sprintf("%s %s (%s): %s", t.TargetProtocol, t.TargetAddress, t.TargetName, t.Summary["handshake"])
}

// ProcessOoniHistorical process historical OONI data by iterating from the
// given "from date" to the given "to date" in one-day increments.
func ProcessOoniHistorical(fromDateStr, toDateStr string) error {

	log.Printf("Incorporating OONI's bridge measurement results from %s to %s.", fromDateStr, toDateStr)
	fromDate, err := time.Parse("2006-01-02", fromDateStr)
	if err != nil {
		return err
	}
	toDate, err := time.Parse("2006-01-02", toDateStr)
	if err != nil {
		return err
	}

	for ; fromDate.Before(toDate); fromDate = fromDate.Add(time.Hour * 24) {
		ProcessOoniDate(fromDate.Format("2006-01-02"))
	}

	return nil
}

// ProcessOoniDate continuously fetches OONI's latest bridge measurement
// results and incorporates them into our database.
func ProcessOoniOnline() {

	ticker := time.NewTicker(OoniDataReloadInterval)
	defer ticker.Stop()

	for ; true; <-ticker.C {
		log.Println("Incorporating OONI's bridge measurement results.")

		// Use yesterday's UTC date.
		yesterday := time.Now().UTC().Unix() - (60 * 60 * 24)
		log.Printf("Time: %s", time.Unix(yesterday, 0).UTC().Format("2006-01-02"))
		ProcessOoniDate(time.Unix(yesterday, 0).UTC().Format("2006-01-02"))
	}
}

// ProcessOoniDate fetches OONI's bridge measurement results for the given date
// and incorporates them into our database.
func ProcessOoniDate(date string) error {
	buckets, err := fetchOoniBuckets(date)
	if err != nil {
		return err
	}

	db, err := sql.Open("sqlite3", config.SqliteFile)
	if err != nil {
		log.Printf("Failed to open SQLite database: %s", err)
		return nil, err
	}
	defer db.Close()

	for _, bucket := range buckets {
		m, err := fetchOoniMeasurement(bucket.FileName)
		if err != nil {
			return err
		}
		if err = augmentSqliteDb(m, db); err != nil {
			return err
		}
	}

	return nil
}

// augmentSqliteDb adds a single OONI measurement to our SQLite database.
func augmentSqliteDb(m *OoniMeasurement, db *sql.DB) error {

	if m.TestName != TorTestName {
		return fmt.Errorf("expected test_name to be %s but got %s", TorTestName, m.TestName)
	}
	log.Printf("Attempting to write OONI measurement to SQLite database.")

	for id, target := range m.TestKeys.Targets {
		if bridge, ok := bridges[id]; ok {
			// We're dealing with a bridge that we know.
			log.Println(bridge)
			// TODO: Add information to our SQL table.
			InsertBlockedBridge(db)

			// SQL schema of our BlockedBridges table:
			// https://gitlab.torproject.org/tpo/anti-censorship/bridgedb/-/blob/develop/bridgedb/Storage.py#L71
		} else {
			log.Printf("Could not find bridge ID in map: %s", id)
		}
	}

	return nil
}

// fetchOoniMeasurement fetches the given OONI Tor measurement and returns its
// JSON content as a OoniMeasurement struct.
func fetchOoniMeasurement(urlSuffix string) (*OoniMeasurement, error) {

	m := &OoniMeasurement{}

	// Download our *.tar.lz4 file.
	s3Url := fmt.Sprintf("%s/%s", S3BaseUrl, urlSuffix)
	log.Printf("Attempting to fetch and process <%s>.", s3Url)

	resp, err := http.Get(s3Url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Write tarball to a temporary file.
	tmpfile, err := ioutil.TempFile("", "ooni-tor-measurement-*.tar.lz4")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpfile.Name())

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if _, err := tmpfile.Write(body); err != nil {
		return nil, err
	}

	// archiver's Walk function transparently decompresses files.
	err = archiver.Walk(tmpfile.Name(), func(f archiver.File) error {
		content, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}
		if err = json.Unmarshal(content, &m); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return m, nil
}

// fetchOoniBuckets fetches OONI's raw index file for the given date and
// returns the measurement "buckets" it found in them.  These buckets contain
// measurements for OONI's Tor tests, which we're interested in.
func fetchOoniBuckets(date string) ([]OoniBucket, error) {

	s3Url := fmt.Sprintf("%s/%s/index.json.gz", S3BaseUrl, date)
	log.Printf("Attempting to fetch and process <%s>.", s3Url)

	resp, err := http.Get(s3Url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	reader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, reader)
	if err != nil {
		return nil, err
	}

	var buckets []OoniBucket
	dec := json.NewDecoder(&buf)
	for dec.More() {
		var bucket OoniBucket
		if err := dec.Decode(&bucket); err != nil {
			return nil, err
		}
		// We're only interested in OONI's Tor test and ignore everything else.
		// A tor test has a file name like "2020-04-01/tor.0.tar.lz4".
		if strings.Contains(bucket.FileName, "/tor.") {
			buckets = append(buckets, bucket)
		}
	}

	return buckets, nil
}
