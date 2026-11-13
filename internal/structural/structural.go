// Package structural turns configuration files into positioned key/value
// assignments (ADR-0005). The engine evaluates assignment rules on them.
package structural

import (
	"path"
	"strconv"
	"strings"
)

// Assignment is one key/value pair with the byte span of its value.
type Assignment struct {
	// KeyPath is the dotted path, e.g. "database.credentials.password".
	// Sequence elements use their index: "servers.0.token".
	KeyPath string
	// Key is the last key segment.
	Key   string
	Value string
	// Start and End are byte offsets of the value in the content. For a
	// quoted value they exclude the quotes.
	Start, End int
}

// Format names an extractor.
type Format string

// Formats.
const (
	FormatNone       Format = ""
	FormatYAML       Format = "yaml"
	FormatJSON       Format = "json"
	FormatDotenv     Format = "dotenv"
	FormatProperties Format = "properties"
	FormatINI        Format = "ini"
	FormatHCL        Format = "hcl"
)

// Detect picks the extractor for a file name. FormatNone means the file is
// unstructured text.
func Detect(name string) Format {
	base := strings.ToLower(path.Base(strings.ReplaceAll(name, "\\", "/")))
	switch ext := path.Ext(base); {
	case ext == ".yaml" || ext == ".yml":
		return FormatYAML
	case ext == ".json":
		return FormatJSON
	case base == ".env" || strings.HasPrefix(base, ".env.") || ext == ".env":
		return FormatDotenv
	case ext == ".properties":
		return FormatProperties
	case ext == ".ini" || ext == ".cfg":
		return FormatINI
	case ext == ".tf" || ext == ".tfvars" || ext == ".hcl":
		return FormatHCL
	}
	return FormatNone
}

// Extract returns the assignments of a structured file. It returns an error
// when the content does not parse; the caller then treats the file as
// unstructured text.
func Extract(f Format, content []byte) ([]Assignment, error) {
	switch f {
	case FormatYAML:
		return extractYAML(content)
	case FormatJSON:
		return extractJSON(content)
	case FormatDotenv:
		return extractDotenv(content), nil
	case FormatProperties:
		return extractProperties(content, false), nil
	case FormatINI:
		return extractProperties(content, true), nil
	case FormatHCL:
		return extractHCL(content), nil
	case FormatNone:
	}
	return nil, nil
}

// joinPath appends a key segment to a dotted path.
func joinPath(prefix []string, key string) (string, []string) {
	segs := append(append([]string{}, prefix...), key)
	return strings.Join(segs, "."), segs
}

func indexKey(i int) string { return strconv.Itoa(i) }
