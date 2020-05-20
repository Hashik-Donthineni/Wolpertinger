package main

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Discard log messages from error handlers.
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

func TestLoadBridgesFromSql(t *testing.T) {

	config.SqliteFile = "/this/database/does/not/exist"
	_, err := loadBridgesFromSql()
	if err == nil {
		t.Error("failed to reject non-existing database")
	}

	config.SqliteFile = "/dev/zero"
	_, err = loadBridgesFromSql()
	if err == nil {
		t.Error("failed to reject bogus database")
	}
}

func TestLoadBridgesFromExtrainfo(t *testing.T) {

	config.ExtrainfoFile = "/this/extrainfo/file/does/not/exist"
	_, err := loadBridgesFromExtrainfo()
	if err == nil {
		t.Error("failed to reject non-existing extrainfo file")
	}

	config.ExtrainfoFile = "/dev/zero"
	_, err = loadBridgesFromExtrainfo()
	if err == nil {
		t.Error("failed to reject bogus extrainfo file")
	}
}
