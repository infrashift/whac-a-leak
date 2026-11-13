package source

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
)

// Commit identifies the commit a history change comes from.
type Commit struct {
	// Seq orders commits from oldest (0) to newest.
	Seq                int
	Hash, Author, Date string
}

// AddedLine is one line that a diff adds.
type AddedLine struct {
	// Line is the 1-based line number in the new version of the file.
	Line int
	Text []byte
}

// FileDiff is the added content of one file in one diff.
type FileDiff struct {
	Commit Commit
	// Path is the new path, relative to the repository root, with slashes.
	Path   string
	Binary bool
	Added  []AddedLine
}

// commitMarker starts the header line that `git log --format` writes for
// each commit. Diff content lines always start with ' ', '+', '-' or '\',
// so the marker cannot collide with content.
const commitMarker = "\x1eCOMMIT\x1f"

// commitFormat is the --format argument that produces commit header lines.
const commitFormat = "--format=%x1eCOMMIT%x1f%H%x1f%an%x1f%aI"

type diffParser struct {
	emit   func(FileDiff) error
	commit Commit
	seq    int
	cur    *FileDiff
	// Remaining lines of the current hunk. While they are positive, every
	// line is hunk content, even if it looks like a header.
	oldLeft, newLeft int
	nextLine         int
}

// parseDiff reads unified diff output (`git diff` or `git log -p`, both with
// -U0) and calls emit once per file with its added lines. It streams: it
// keeps only the current file in memory.
func parseDiff(r io.Reader, emit func(FileDiff) error) error {
	p := &diffParser{emit: emit, seq: -1}
	br := bufio.NewReaderSize(r, 64<<10)
	for {
		line, err := br.ReadBytes('\n')
		if len(line) > 0 {
			if perr := p.line(bytes.TrimSuffix(line, []byte{'\n'})); perr != nil {
				return perr
			}
		}
		if errors.Is(err, io.EOF) {
			return p.flush()
		}
		if err != nil {
			return err
		}
	}
}

func (p *diffParser) flush() error {
	cur := p.cur
	p.cur = nil
	if cur == nil || cur.Path == "" || cur.Binary || len(cur.Added) == 0 {
		return nil
	}
	return p.emit(*cur)
}

func (p *diffParser) line(l []byte) error {
	if p.oldLeft > 0 || p.newLeft > 0 {
		switch {
		case len(l) > 0 && l[0] == '+' && p.newLeft > 0:
			p.cur.Added = append(p.cur.Added, AddedLine{Line: p.nextLine, Text: bytes.Clone(l[1:])})
			p.nextLine++
			p.newLeft--
			return nil
		case len(l) > 0 && l[0] == '-' && p.oldLeft > 0:
			p.oldLeft--
			return nil
		case len(l) > 0 && l[0] == ' ':
			p.oldLeft--
			p.newLeft--
			p.nextLine++
			return nil
		case len(l) > 0 && l[0] == '\\':
			return nil // "\ No newline at end of file"
		}
		// A malformed count: fall through and read the line as a header.
		p.oldLeft, p.newLeft = 0, 0
	}

	switch {
	case bytes.HasPrefix(l, []byte(commitMarker)):
		if err := p.flush(); err != nil {
			return err
		}
		parts := bytes.SplitN(l[len(commitMarker):], []byte{0x1f}, 3)
		p.seq++
		p.commit = Commit{Seq: p.seq}
		if len(parts) == 3 {
			p.commit.Hash, p.commit.Author, p.commit.Date = string(parts[0]), string(parts[1]), string(bytes.TrimSpace(parts[2]))
		}
	case bytes.HasPrefix(l, []byte("diff --git ")):
		if err := p.flush(); err != nil {
			return err
		}
		p.cur = &FileDiff{Commit: p.commit}
	case p.cur == nil:
		// Commit message lines or other text before the first file.
	case bytes.HasPrefix(l, []byte("Binary files ")) || bytes.HasPrefix(l, []byte("GIT binary patch")):
		p.cur.Binary = true
	case bytes.HasPrefix(l, []byte("+++ ")):
		path, err := diffPath(l[4:])
		if err != nil {
			return err
		}
		p.cur.Path = path
	case bytes.HasPrefix(l, []byte("@@ ")):
		return p.hunk(l)
	}
	return nil
}

// hunk parses "@@ -a[,b] +c[,d] @@".
func (p *diffParser) hunk(l []byte) error {
	fields := bytes.Fields(l)
	if len(fields) < 3 {
		return fmt.Errorf("malformed hunk header %q", l)
	}
	_, oldCount, err := hunkRange(fields[1], '-')
	if err != nil {
		return err
	}
	newStart, newCount, err := hunkRange(fields[2], '+')
	if err != nil {
		return err
	}
	p.oldLeft, p.newLeft, p.nextLine = oldCount, newCount, newStart
	return nil
}

func hunkRange(f []byte, sign byte) (start, count int, err error) {
	if len(f) < 2 || f[0] != sign {
		return 0, 0, fmt.Errorf("malformed hunk range %q", f)
	}
	s, c, found := bytes.Cut(f[1:], []byte{','})
	if start, err = strconv.Atoi(string(s)); err != nil {
		return 0, 0, fmt.Errorf("malformed hunk range %q", f)
	}
	count = 1
	if found {
		if count, err = strconv.Atoi(string(c)); err != nil {
			return 0, 0, fmt.Errorf("malformed hunk range %q", f)
		}
	}
	return start, count, nil
}

// diffPath decodes the path of a "+++ " line: "b/path", a C-quoted
// "\"b/pa\\th\"", or /dev/null for a deleted file (returned as "").
func diffPath(raw []byte) (string, error) {
	s := string(bytes.TrimRight(raw, "\t\r"))
	if s == "/dev/null" {
		return "", nil
	}
	if s != "" && s[0] == '"' {
		u, err := unquoteC(s)
		if err != nil {
			return "", fmt.Errorf("diff path %s: %w", s, err)
		}
		s = u
	}
	if len(s) < 2 || s[:2] != "b/" {
		return "", fmt.Errorf("diff path %q lacks the b/ prefix", s)
	}
	return s[2:], nil
}

// unquoteC decodes git's C-style quoting: \a \b \t \n \v \f \r \" \\ and
// three-digit octal escapes for raw bytes.
func unquoteC(s string) (string, error) {
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return "", errors.New("unbalanced quotes")
	}
	in := s[1 : len(s)-1]
	var out []byte
	for i := 0; i < len(in); i++ {
		c := in[i]
		if c != '\\' {
			out = append(out, c)
			continue
		}
		i++
		if i >= len(in) {
			return "", errors.New("dangling escape")
		}
		switch e := in[i]; e {
		case 'a':
			out = append(out, '\a')
		case 'b':
			out = append(out, '\b')
		case 't':
			out = append(out, '\t')
		case 'n':
			out = append(out, '\n')
		case 'v':
			out = append(out, '\v')
		case 'f':
			out = append(out, '\f')
		case 'r':
			out = append(out, '\r')
		case '"', '\\':
			out = append(out, e)
		default:
			if i+2 >= len(in) {
				return "", fmt.Errorf("bad escape \\%c", e)
			}
			v, err := strconv.ParseUint(in[i:i+3], 8, 8)
			if err != nil {
				return "", fmt.Errorf("bad escape \\%s", in[i:i+3])
			}
			out = append(out, byte(v))
			i += 2
		}
	}
	return string(out), nil
}
