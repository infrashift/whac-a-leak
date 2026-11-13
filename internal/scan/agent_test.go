package scan

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/infrashift/whac-a-leak/internal/rules"
)

func TestAgentContext(t *testing.T) {
	home := t.TempDir()
	for rel, content := range map[string]string{ //nolint:gosec // G101: fake credential files with EXAMPLE markers
		".aws/credentials":       "[default]\naws_access_key_id = AKIAIOSFODNN7EXAMPLE\n",
		".aws/sso/cache/x.json":  `{"accessToken": "EXAMPLEaB3dE5fG7hJ9kL2mN4pQ6rS8tU"}`,
		".kube/config":           "users:\n- user:\n    client-key-data: RVhBTVBMRS1rZXktZGF0YS1ub3QtcmVhbA==\n",
		".netrc":                 "machine api.example.com login app password EXAMPLE-hunter2\n",
		".config/unrelated.json": `{"accessToken": "EXAMPLEaB3dE5fG7hJ9kL2mN4pQ6rS8tU"}`,
	} {
		p := filepath.Join(home, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	p := pipeline(t, nil)
	res, err := p.AgentContext(context.Background(), &AgentOptions{
		Env: true, Home: true,
		Environ: []string{"PATH=/usr/bin", "GH_TOKEN=" + token, "DB_PASSWORD=EXAMPLE-letmein", "EMPTY_TOKEN="},
		HomeDir: home,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, f := range res.Findings {
		got[f.RuleID+" "+f.Path+" "+f.KeyPath] = true
		if f.Secret != "" || f.Redacted != "[masked]" {
			t.Errorf("value not masked: %+v", f)
		}
		if strings.Contains(f.Path, "unrelated") {
			t.Errorf("a file outside the credential paths was read: %+v", f)
		}
	}
	for _, want := range []string{
		"GITHUB_TOKEN env env.GH_TOKEN",
		"SENSITIVE_PATH env env.DB_PASSWORD",
		"AWS_ACCESS_KEY_ID ~/.aws/credentials ",
		"SESSION_TOKEN ~/.aws/sso/cache/x.json accessToken",
		"SESSION_TOKEN ~/.kube/config users.0.user.client-key-data",
		"NETRC_PASSWORD ~/.netrc ",
	} {
		if !got[want] {
			t.Errorf("missing finding %q; got %v", want, got)
		}
	}
	if res.Stats.Files != 5 {
		t.Errorf("files = %d, want 5 (env + 4 credential files)", res.Stats.Files)
	}
}

func TestAgentContextExposureAndAllowlist(t *testing.T) {
	p := pipeline(t, func(rs *rules.Ruleset) {
		rs.Allowlist.Paths["/allowed/**"] = "accepted"
	})
	res, err := p.AgentContext(context.Background(), &AgentOptions{
		Exposure: true,
		Getenv:   func(string) string { return "" },
	})
	if err != nil || len(res.Findings) != 0 {
		t.Errorf("nothing reachable: %+v, %v", res, err)
	}
}
