package engine

import (
	"regexp"
	"strings"
)

var (
	uuidRe    = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	pathRe    = regexp.MustCompile(`^(?:~|\.{1,2})?/(?:[A-Za-z0-9._-]+/)*[A-Za-z0-9._-]+/?$|^[A-Za-z]:\\`)
	urlRe     = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9+.-]*://([^/?#]*)`)
	hexRe     = regexp.MustCompile(`^(?:[0-9a-fA-F]{32}|[0-9a-fA-F]{40}|[0-9a-fA-F]{56}|[0-9a-fA-F]{64}|[0-9a-fA-F]{96}|[0-9a-fA-F]{128})$`)
	hashKeyRe = regexp.MustCompile(`(?i)(sha|md5|hash|digest|checksum|integrity|etag|commit|rev)`)
)

// safeFormat returns why a high-entropy candidate is a known non-secret
// format, or "". key is the assignment key, or "" for a pattern match. The
// engine applies it only to rules with an entropy block (ADR-0006).
func safeFormat(secret, key string) string {
	switch {
	case uuidRe.MatchString(secret):
		return "secret is a UUID"
	case pathRe.MatchString(secret):
		return "secret is a file path"
	case hashKeyRe.MatchString(key) && hexRe.MatchString(secret):
		return "secret is a hash under a hash-like key"
	}
	if m := urlRe.FindStringSubmatch(secret); len(m) == 2 && !strings.Contains(m[1], "@") {
		return "secret is a URL without credentials"
	}
	return ""
}
