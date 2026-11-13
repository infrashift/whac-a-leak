package source

import (
	"strings"
	"testing"
)

const sampleLog = "\x1eCOMMIT\x1faaa111\x1fAda\x1f2026-01-01T00:00:00Z\n" +
	"\n" +
	"diff --git a/app.env b/app.env\n" +
	"new file mode 100644\n" +
	"index 0000000..1111111\n" +
	"--- /dev/null\n" +
	"+++ b/app.env\n" +
	"@@ -0,0 +1,3 @@\n" +
	"+A=1\n" +
	"++++ not a header: an added line that starts with ++\n" +
	"+C=3\r\n" +
	"diff --git a/logo.png b/logo.png\n" +
	"Binary files /dev/null and b/logo.png differ\n" +
	"\x1eCOMMIT\x1fbbb222\x1fBo\x1f2026-01-02T00:00:00Z\n" +
	"diff --git a/app.env b/app.env\n" +
	"--- a/app.env\n" +
	"+++ b/app.env\n" +
	"@@ -1 +1 @@\n" +
	"-A=1\n" +
	"+A=2\n" +
	"\\ No newline at end of file\n" +
	"@@ -5,0 +6,2 @@\n" +
	"+F=6\n" +
	"+G=7\n" +
	"diff --git a/old.txt b/old.txt\n" +
	"deleted file mode 100644\n" +
	"--- a/old.txt\n" +
	"+++ /dev/null\n" +
	"@@ -1 +0,0 @@\n" +
	"-gone\n" +
	"diff --git \"a/sp ace\\tx.env\" \"b/sp ace\\tx.env\"\n" +
	"--- /dev/null\n" +
	"+++ \"b/sp ace\\tx.env\"\n" +
	"@@ -0,0 +1 @@\n" +
	"+Q=1\n"

func TestParseDiff(t *testing.T) {
	var got []FileDiff
	if err := parseDiff(strings.NewReader(sampleLog), func(d FileDiff) error { got = append(got, d); return nil }); err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d file diffs, want 3 (binary and deleted files dropped): %+v", len(got), got)
	}
	first := got[0]
	if first.Path != "app.env" || first.Commit.Hash != "aaa111" || first.Commit.Seq != 0 || len(first.Added) != 3 {
		t.Fatalf("first = %+v", first)
	}
	if string(first.Added[1].Text) != "+++ not a header: an added line that starts with ++" || first.Added[1].Line != 2 {
		t.Errorf("hunk content looked like a header: %+v", first.Added[1])
	}
	if string(first.Added[2].Text) != "C=3\r" {
		t.Errorf("CRLF line = %q", first.Added[2].Text)
	}
	second := got[1]
	if second.Commit.Author != "Bo" || second.Commit.Seq != 1 || len(second.Added) != 3 ||
		second.Added[0].Line != 1 || second.Added[1].Line != 6 || second.Added[2].Line != 7 {
		t.Errorf("second = %+v", second)
	}
	if got[2].Path != "sp ace\tx.env" {
		t.Errorf("quoted path = %q", got[2].Path)
	}
}

func TestParseDiffErrors(t *testing.T) {
	bad := []string{
		"diff --git a/x b/x\n+++ b/x\n@@ -1 +x @@\n",
		"diff --git a/x b/x\n+++ nob/x\n",
		"diff --git a/x b/x\n+++ \"b/unterminated\n",
	}
	for _, in := range bad {
		if err := parseDiff(strings.NewReader(in), func(FileDiff) error { return nil }); err == nil {
			t.Errorf("parseDiff(%q): want an error", in)
		}
	}
}

func TestUnquoteC(t *testing.T) {
	got, err := unquoteC(`"a\"b\\c\303\251"`)
	if err != nil || got != "a\"b\\cé" {
		t.Errorf("unquoteC = %q, %v", got, err)
	}
	for _, bad := range []string{`"a`, `"a\"`, `"\9"`} {
		if _, err := unquoteC(bad); err == nil {
			t.Errorf("unquoteC(%q): want an error", bad)
		}
	}
}
