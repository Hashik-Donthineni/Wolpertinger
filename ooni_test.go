package main

import (
	"fmt"
	"testing"
)

func TestFetchOoniBuckets(t *testing.T) {

	validDate := "2020-05-01"
	buckets, err := fetchOoniBuckets(validDate)
	if err != nil {
		t.Error("couldn't fetch data for valid date")
	}
	if len(buckets) != 1 {
		t.Error("found more than just one bucket")
	}
	if buckets[0].FileName != fmt.Sprintf("%s/tor.0.tar.lz4", validDate) {
		t.Error("found incorrect file name")
	}

	_, err = fetchOoniBuckets("1970-01-01")
	if err == nil {
		t.Error("somehow wound up with data for incorrect date")
	}
}

func TestFetchOoniMeasurement(t *testing.T) {

	_, err := fetchOoniMeasurement("2020-05-01/tor.0.tar.lz4")
	if err != nil {
		t.Error("failed to fetch valid measurement")
	}

	_, err = fetchOoniMeasurement("2020-05-01/tor.1.tar.lz4")
	if err == nil {
		t.Error("somehow wound up with data for non-existing tarball")
	}
}
