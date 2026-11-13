package structural

import (
	"bytes"
	"regexp"
	"strings"
)

// forEachLine calls fn with each line and the byte offset of its start.
func forEachLine(content []byte, fn func(line []byte, off int)) {
	off := 0
	for off <= len(content) {
		end := bytes.IndexByte(content[off:], '\n')
		if end < 0 {
			fn(content[off:], off)
			return
		}
		fn(content[off:off+end], off)
		off += end + 1
	}
}

// unquote returns the value inside matching quotes and its offset shift, or
// the trimmed bare value with an inline comment removed.
func unquote(raw, comment string) (value string, shift int) {
	lead := len(raw) - len(strings.TrimLeft(raw, " \t"))
	s := raw[lead:]
	if len(s) >= 1 && (s[0] == '"' || s[0] == '\'') {
		if end := strings.IndexByte(s[1:], s[0]); end >= 0 {
			return s[1 : 1+end], lead + 1
		}
	}
	if comment != "" {
		if i := strings.Index(s, comment); i >= 0 {
			s = s[:i]
		}
	}
	return strings.TrimRight(s, " \t\r"), lead
}

var dotenvRe = regexp.MustCompile(`^\s*(?:export\s+)?([A-Za-z_][A-Za-z0-9_.-]*)\s*=(.*)$`)

func extractDotenv(content []byte) []Assignment {
	var out []Assignment
	forEachLine(content, func(line []byte, off int) {
		m := dotenvRe.FindSubmatchIndex(line)
		if m == nil {
			return
		}
		key := string(line[m[2]:m[3]])
		value, shift := unquote(string(line[m[4]:m[5]]), " #")
		start := off + m[4] + shift
		out = append(out, Assignment{KeyPath: key, Key: key, Value: value, Start: start, End: start + len(value)})
	})
	return out
}

var (
	sectionRe  = regexp.MustCompile(`^\s*\[([^\]]+)\]\s*$`)
	propertyRe = regexp.MustCompile(`^\s*([^\s=:#;!][^=:]*?)\s*[=:]\s?(.*)$`)
)

// extractProperties reads Java properties (ini=false) or INI files
// (ini=true, with [section] prefixes).
func extractProperties(content []byte, ini bool) []Assignment {
	var (
		out     []Assignment
		section string
	)
	forEachLine(content, func(line []byte, off int) {
		trimmed := strings.TrimSpace(string(line))
		if trimmed == "" || strings.ContainsAny(trimmed[:1], "#;!") {
			return
		}
		if ini {
			if m := sectionRe.FindStringSubmatch(trimmed); m != nil {
				section = strings.TrimSpace(m[1])
				return
			}
		}
		m := propertyRe.FindSubmatchIndex(line)
		if m == nil {
			return
		}
		key := strings.TrimSpace(string(line[m[2]:m[3]]))
		value, shift := unquote(string(line[m[4]:m[5]]), "")
		keyPath := key
		if section != "" {
			keyPath = section + "." + key
		}
		last := keyPath[strings.LastIndexByte(keyPath, '.')+1:]
		start := off + m[4] + shift
		out = append(out, Assignment{KeyPath: keyPath, Key: last, Value: value, Start: start, End: start + len(value)})
	})
	return out
}

var (
	hclBlockRe = regexp.MustCompile(`^\s*([A-Za-z_][\w-]*)((?:\s+"[^"]*")*)\s*\{\s*$`)
	hclLabelRe = regexp.MustCompile(`"([^"]*)"`)
	hclAttrRe  = regexp.MustCompile(`^\s*([A-Za-z_][\w-]*)\s*=\s*"((?:[^"\\]|\\.)*)"`)
	hclCloseRe = regexp.MustCompile(`^\s*\}`)
)

// extractHCL reads HCL-style `key = "value"` attributes inside nested
// blocks. Unquoted values are references or expressions and are skipped.
func extractHCL(content []byte) []Assignment {
	var (
		out   []Assignment
		stack []string
	)
	forEachLine(content, func(line []byte, off int) {
		s := string(line)
		switch {
		case hclBlockRe.MatchString(s):
			m := hclBlockRe.FindStringSubmatch(s)
			labels := hclLabelRe.FindAllStringSubmatch(m[2], -1)
			parts := make([]string, 0, 1+len(labels))
			parts = append(parts, m[1])
			for _, l := range labels {
				parts = append(parts, l[1])
			}
			stack = append(stack, strings.Join(parts, "."))
		case hclCloseRe.MatchString(s):
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		default:
			m := hclAttrRe.FindStringSubmatchIndex(s)
			if m == nil {
				return
			}
			key := s[m[2]:m[3]]
			keyPath, _ := joinPath(stack, key)
			start := off + m[4]
			out = append(out, Assignment{KeyPath: keyPath, Key: key, Value: s[m[4]:m[5]], Start: start, End: off + m[5]})
		}
	})
	return out
}

// genericRe finds `key = value`, `key: value`, `key := value` and
// `key => value` pairs in source code and other unstructured text.
var genericRe = regexp.MustCompile(
	`(?:^|[\s{(,;\[])["']?([A-Za-z_][A-Za-z0-9_.\-]*)["']?\s*(:=|=>|==|===|!=|=|:)\s*` +
		`(?:"((?:[^"\\\n]|\\.)*)"|'((?:[^'\\\n]|\\.)*)'|` + "`" + `([^` + "`" + `\n]*)` + "`" +
		`|([^\s"'` + "`" + `,;#(){}\[\]]+))`)

// Generic extracts key/value pairs from one window of unstructured text.
// base is the byte offset of the window in the content.
func Generic(window []byte, base int) []Assignment {
	var out []Assignment
	for _, m := range genericRe.FindAllSubmatchIndex(window, -1) {
		if op := string(window[m[4]:m[5]]); op == "==" || op == "===" || op == "!=" {
			continue
		}
		vs, ve := -1, -1
		for g := 3; g <= 6; g++ {
			if m[2*g] >= 0 {
				vs, ve = m[2*g], m[2*g+1]
				break
			}
		}
		if vs < 0 {
			continue
		}
		keyPath := string(window[m[2]:m[3]])
		key := keyPath[strings.LastIndexByte(keyPath, '.')+1:]
		out = append(out, Assignment{KeyPath: keyPath, Key: key, Value: string(window[vs:ve]), Start: base + vs, End: base + ve})
	}
	return out
}
