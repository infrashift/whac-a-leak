package engine

import "regexp"

var uuidRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// safeFormat returns why a high-entropy candidate is a known non-secret
// format, or "". Milestone 3 adds formats that need key context, such as
// hashes under checksum or integrity keys.
func safeFormat(secret string) string {
	if uuidRe.MatchString(secret) {
		return "secret is a UUID"
	}
	return ""
}
