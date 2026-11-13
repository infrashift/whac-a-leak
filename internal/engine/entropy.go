// Package engine is Whac-a-Leak's scan engine: plain Go matching over rules
// compiled by package rules. It never touches the CUE runtime (ADR-0003).
package engine

import (
	"fmt"
	"math"

	"github.com/infrashift/whac-a-leak/internal/rules"
)

// defaultMinFraction is the share of the maximum reachable entropy a secret
// must have when a rule sets no explicit minimum. The maximum depends on both
// the alphabet and the length: a string of n characters drawn from k symbols
// can reach at most log2(min(n, k)) bits per character (ADR-0006).
const defaultMinFraction = 0.75

// Shannon returns the Shannon entropy of s in bits per byte.
func Shannon(s string) float64 {
	if s == "" {
		return 0
	}
	var counts [256]int
	for i := 0; i < len(s); i++ {
		counts[s[i]]++
	}
	n := float64(len(s))
	var h float64
	for _, c := range &counts {
		if c == 0 {
			continue
		}
		p := float64(c) / n
		h -= p * math.Log2(p)
	}
	return h
}

// Classify returns the smallest charset that contains every byte of s.
func Classify(s string) rules.Charset {
	hex, alnum, b64url, b64 := true, true, true, true
	for i := 0; i < len(s); i++ {
		c := s[i]
		isDigit := c >= '0' && c <= '9'
		isAlpha := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
		isHex := isDigit || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
		hex = hex && isHex
		alnum = alnum && (isDigit || isAlpha)
		b64url = b64url && (isDigit || isAlpha || c == '-' || c == '_')
		b64 = b64 && (isDigit || isAlpha || c == '+' || c == '/' || c == '=')
	}
	switch {
	case hex:
		return rules.CharsetHex
	case alnum:
		return rules.CharsetAlnum
	case b64url:
		return rules.CharsetBase64URL
	case b64:
		return rules.CharsetBase64
	}
	return rules.CharsetAny
}

// alphabetSize is the number of distinct symbols in a charset.
func alphabetSize(c rules.Charset) float64 {
	switch c {
	case rules.CharsetHex:
		return 16
	case rules.CharsetAlnum:
		return 62
	case rules.CharsetBase64, rules.CharsetBase64URL:
		return 64
	case rules.CharsetAny:
		return 95 // printable ASCII
	}
	return 95
}

// MinEntropy is the threshold a secret must reach under e: the explicit
// minimum when set, otherwise a fraction of the maximum entropy reachable for
// the secret's length and charset.
func MinEntropy(e *rules.Entropy, secret string) float64 {
	if e.Min != nil {
		return *e.Min
	}
	cs := e.Charset
	if cs == "" {
		cs = Classify(secret)
	}
	reachable := math.Min(float64(len(secret)), alphabetSize(cs))
	if reachable < 2 {
		return 0
	}
	return defaultMinFraction * math.Log2(reachable)
}

// checkEntropy returns why secret fails e, or "" if it passes.
func checkEntropy(e *rules.Entropy, secret string) string {
	if len(secret) < e.MinLen {
		return fmt.Sprintf("secret shorter than minLen %d", e.MinLen)
	}
	h := Shannon(secret)
	if minH := MinEntropy(e, secret); h < minH {
		return fmt.Sprintf("entropy %.2f below minimum %.2f", h, minH)
	}
	if e.Max != nil && h > *e.Max {
		return fmt.Sprintf("entropy %.2f above maximum %.2f", h, *e.Max)
	}
	return ""
}
