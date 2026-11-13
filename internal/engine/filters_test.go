package engine

import "testing"

func TestPlaceholder(t *testing.T) {
	placeholders := []string{
		"", "  ", "${DB_PASSWORD}", "prefix-${X}", "{{ .Values.token }}", "$API_TOKEN", "%TOKEN%",
		"<your-token>", "changeme", "CHANGEME", "password", "your_api_key_here", "insert-token",
		"sk-xxxxxxxxxxxx", "ab****yz", "aaaaaa", "000000",
	}
	for _, v := range placeholders {
		if placeholder(v) == "" {
			t.Errorf("placeholder(%q) = \"\", want a reason", v)
		}
	}
	secrets := []string{"EXAMPLE-hunter2", "Zx9EXAMPLEq7Lm2Vt8Rk4Wp6", "p@ss word", "aab", "$5ecret"}
	for _, v := range secrets {
		if r := placeholder(v); r != "" {
			t.Errorf("placeholder(%q) = %q, want \"\"", v, r)
		}
	}
}

func TestSafeFormat(t *testing.T) {
	const sha = "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"
	tests := []struct {
		secret, key string
		safe        bool
	}{
		{"550e8400-e29b-41d4-a716-446655440000", "", true},
		{"/var/run/secrets/token", "token_file", true},
		{"./certs/client.key", "key", true},
		{"~/.config/app/token", "token", true},
		{`C:\Users\app\token.txt`, "token", true},
		{"https://auth.example.com/oauth/token", "auth_url", true},
		{"https://user:EXAMPLEpass@db.example.com/app", "url", false},
		{sha, "tokenSha256", true},
		{sha, "token", false},
		{"Zx9EXAMPLEq7Lm2Vt8Rk4Wp6", "token", false},
		{"ab/cd+ef/gh==", "token", false},
	}
	for _, tt := range tests {
		if got := safeFormat(tt.secret, tt.key) != ""; got != tt.safe {
			t.Errorf("safeFormat(%q, %q) safe = %v, want %v", tt.secret, tt.key, got, tt.safe)
		}
	}
}

func TestKeyPathForGlob(t *testing.T) {
	if got := keyPathForGlob("Database.Credentials.Password"); got != "database/credentials/password" {
		t.Errorf("got %q", got)
	}
	if got := keyPathForGlob("labels.app.kubernetes.io/name"); got != "labels/app/kubernetes/io%2fname" {
		t.Errorf("slash in a segment: %q", got)
	}
}
