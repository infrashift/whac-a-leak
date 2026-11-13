package fingerprint

import "testing"

func TestOf(t *testing.T) {
	// Test vector: printf 'RULE\0secret' | sha256sum
	const want = "whac1:3705e2061f9427e02a737f19dde21c8e2db1bc6aa8f1c1c24fc2ce09a1faabee"
	if got := Of("RULE", "secret"); got != want {
		t.Errorf("Of = %s, want %s", got, want)
	}
	if !Valid(want) {
		t.Error("Valid rejected a well-formed fingerprint")
	}
}

func TestOfIsStableAndSeparated(t *testing.T) {
	if Of("R", "abc") != Of("R", ` "abc" `) {
		t.Error("quotes and surrounding space must not change the fingerprint")
	}
	if Of("R", "abc") == Of("R", "ABC") {
		t.Error("case must change the fingerprint")
	}
	if Of("AB", "C") == Of("A", "BC") {
		t.Error("the separator must keep the rule ID and secret apart")
	}
	if Of("R1", "abc") == Of("R2", "abc") {
		t.Error("the rule ID must change the fingerprint")
	}
}

func TestNormalize(t *testing.T) {
	tests := map[string]string{
		"  abc  ": "abc",
		`"abc"`:   "abc",
		"'abc'":   "abc",
		"`abc`":   "abc",
		`"abc'`:   `"abc'`,
		`"`:       `"`,
		`""`:      "",
	}
	for in, want := range tests {
		if got := Normalize(in); got != want {
			t.Errorf("Normalize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestValid(t *testing.T) {
	if Valid("whac1:abc") || Valid("wal1:"+string(make([]byte, 64))) {
		t.Error("Valid accepted a malformed fingerprint")
	}
}
