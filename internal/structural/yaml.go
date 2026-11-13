package structural

import (
	"bytes"
	"errors"
	"io"
	"unicode/utf8"

	"go.yaml.in/yaml/v3"
)

func extractYAML(content []byte) ([]Assignment, error) {
	lines := lineStarts(content)
	dec := yaml.NewDecoder(bytes.NewReader(content))
	var out []Assignment
	for {
		var doc yaml.Node
		err := dec.Decode(&doc)
		if errors.Is(err, io.EOF) {
			return out, nil
		}
		if err != nil {
			return nil, err
		}
		walkYAML(&doc, nil, func(segs []string, v *yaml.Node) {
			start := offsetOf(content, lines, v.Line, v.Column)
			if v.Style&(yaml.DoubleQuotedStyle|yaml.SingleQuotedStyle) != 0 {
				start++ // the column points at the opening quote
			}
			key := ""
			if len(segs) > 0 {
				key = segs[len(segs)-1]
			}
			end := min(len(content), start+len(v.Value))
			out = append(out, Assignment{KeyPath: joinSegs(segs), Key: key, Value: v.Value, Start: start, End: end})
		})
	}
}

func walkYAML(n *yaml.Node, segs []string, emit func([]string, *yaml.Node)) {
	switch n.Kind {
	case yaml.DocumentNode:
		for _, c := range n.Content {
			walkYAML(c, segs, emit)
		}
	case yaml.MappingNode:
		for i := 0; i+1 < len(n.Content); i += 2 {
			_, next := joinPath(segs, n.Content[i].Value)
			walkYAML(n.Content[i+1], next, emit)
		}
	case yaml.SequenceNode:
		for i, c := range n.Content {
			_, next := joinPath(segs, indexKey(i))
			walkYAML(c, next, emit)
		}
	case yaml.ScalarNode:
		if n.Tag != "!!null" && n.Tag != "!!bool" && len(segs) > 0 {
			emit(segs, n)
		}
	case yaml.AliasNode:
		// An alias repeats a value reported at its anchor.
	}
}

func joinSegs(segs []string) string {
	p, _ := joinPath(segs[:len(segs)-1], segs[len(segs)-1])
	return p
}

// lineStarts returns the byte offset of each line start.
func lineStarts(content []byte) []int {
	starts := []int{0}
	for i, b := range content {
		if b == '\n' {
			starts = append(starts, i+1)
		}
	}
	return starts
}

// offsetOf converts a 1-based line and a 1-based column in characters into
// a byte offset.
func offsetOf(content []byte, starts []int, line, col int) int {
	if line < 1 || line > len(starts) {
		return 0
	}
	off := starts[line-1]
	for c := 1; c < col && off < len(content) && content[off] != '\n'; c++ {
		_, size := utf8.DecodeRune(content[off:])
		off += size
	}
	return off
}
