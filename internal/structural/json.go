package structural

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

type jsonFrame struct {
	object  bool
	key     string
	index   int
	wantKey bool
}

func extractJSON(content []byte) ([]Assignment, error) {
	dec := json.NewDecoder(bytes.NewReader(content))
	dec.UseNumber()
	var (
		out   []Assignment
		stack []jsonFrame
	)
	segs := func() []string {
		s := make([]string, 0, len(stack))
		for _, f := range stack {
			if f.object {
				s = append(s, f.key)
			} else {
				s = append(s, indexKey(f.index))
			}
		}
		return s
	}
	advance := func() {
		if n := len(stack); n > 0 {
			if stack[n-1].object {
				stack[n-1].wantKey = true
			} else {
				stack[n-1].index++
			}
		}
	}
	for {
		before := int(dec.InputOffset())
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			if len(stack) > 0 {
				return nil, io.ErrUnexpectedEOF
			}
			return out, nil
		}
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case json.Delim:
			switch t {
			case '{', '[':
				stack = append(stack, jsonFrame{object: t == '{', wantKey: t == '{'})
			case '}', ']':
				stack = stack[:len(stack)-1]
				advance()
			}
		case string:
			if n := len(stack); n > 0 && stack[n-1].object && stack[n-1].wantKey {
				stack[n-1].key = t
				stack[n-1].wantKey = false
				continue
			}
			if s := segs(); len(s) > 0 {
				start := before + bytes.IndexByte(content[before:], '"') + 1
				end := int(dec.InputOffset()) - 1
				out = append(out, Assignment{KeyPath: joinSegs(s), Key: s[len(s)-1], Value: t, Start: start, End: end})
			}
			advance()
		default:
			advance()
		}
	}
}
