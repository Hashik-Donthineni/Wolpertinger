package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// Hmac calculates and returns a HMAC-SHA256 over the provided data.  The key
// is from our configuration file.
func Hmac(data []byte) string {

	h := hmac.New(sha256.New, []byte(config.MasterKey))
	h.Write([]byte(data))

	return hex.EncodeToString(h.Sum(nil))
}
