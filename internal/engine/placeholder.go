package engine

import (
	"regexp"
	"strings"
)

var placeholderWords = map[string]bool{
	"changeme": true, "change_me": true, "change-me": true, "password": true, "passwd": true,
	"secret": true, "token": true, "todo": true, "tbd": true, "fixme": true, "null": true,
	"nil": true, "none": true, "undefined": true, "redacted": true, "dummy": true,
	"placeholder": true, "default": true,
}

var (
	envRefRe    = regexp.MustCompile(`^\$[A-Za-z_][A-Za-z0-9_]*$|^%[A-Za-z_][A-Za-z0-9_]*%$`)
	angleRe     = regexp.MustCompile(`^<[^<>]*>$`)
	yourValueRe = regexp.MustCompile(`(?i)^(your|insert|enter|replace)[-_ ]`)
)

// placeholder returns why a value is a placeholder rather than a secret, or
// "". Assignment rules apply it before any other value check (ADR-0005).
func placeholder(value string) string {
	s := strings.TrimSpace(value)
	lower := strings.ToLower(s)
	switch {
	case s == "":
		return "value is empty"
	case strings.Contains(s, "${") || strings.Contains(s, "{{") || envRefRe.MatchString(s):
		return "value is a variable reference"
	case angleRe.MatchString(s):
		return "value is a <placeholder>"
	case placeholderWords[lower]:
		return "value is a placeholder word"
	case yourValueRe.MatchString(s):
		return "value is placeholder text"
	case strings.Contains(lower, "xxxx") || strings.Contains(s, "****"):
		return "value is masked"
	case len(s) >= 3 && strings.Count(s, s[:1]) == len(s):
		return "value repeats one character"
	}
	return ""
}
