// Package fingerprint computes location-independent finding fingerprints
// (spec: Fingerprint format; ADR-0007).
package fingerprint

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

// Prefix versions the fingerprint scheme.
const Prefix = "whac1:"

var validRe = regexp.MustCompile(`^whac1:[0-9a-f]{64}$`)

// Normalize trims surrounding whitespace and one pair of matching quotes.
// It does not fold case or normalize Unicode.
func Normalize(secret string) string {
	s := strings.TrimSpace(secret)
	if len(s) >= 2 {
		if q := s[0]; (q == '"' || q == '\'' || q == '`') && s[len(s)-1] == q {
			s = s[1 : len(s)-1]
		}
	}
	return s
}

// Of returns whac1:hex(SHA-256(ruleID || 0x00 || Normalize(secret))). The
// zero byte separates the fields, so ("AB", "C…") and ("A", "BC…") differ.
func Of(ruleID, secret string) string {
	h := sha256.New()
	h.Write([]byte(ruleID))
	h.Write([]byte{0})
	h.Write([]byte(Normalize(secret)))
	return Prefix + hex.EncodeToString(h.Sum(nil))
}

// Valid reports whether fp is a well-formed whac1 fingerprint.
func Valid(fp string) bool { return validRe.MatchString(fp) }
