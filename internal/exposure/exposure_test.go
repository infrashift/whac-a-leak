package exposure

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/infrashift/whac-a-leak/internal/rules"
)

var allOn = rules.Exposure{
	SSHAgent:      rules.ExposureCheck{Severity: rules.SeverityHigh, Enabled: true},
	DockerSocket:  rules.ExposureCheck{Severity: rules.SeverityCritical, Enabled: true},
	CloudMetadata: rules.ExposureCheck{Severity: rules.SeverityHigh, Enabled: true},
}

// listen opens a Unix socket in a short temp path (socket paths are limited
// to about 100 bytes).
func listen(t *testing.T, name string) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "whac") //nolint:usetesting // t.TempDir paths can exceed the Unix socket path limit
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	p := filepath.Join(dir, name)
	var lc net.ListenConfig
	l, err := lc.Listen(t.Context(), "unix", p)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = l.Close() })
	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}
			_ = c.Close()
		}
	}()
	return p
}

func env(m map[string]string) func(string) string { return func(k string) string { return m[k] } }

func ids(fs []Finding) map[string]int {
	out := map[string]int{}
	for _, f := range fs {
		out[f.ID]++
	}
	return out
}

func TestSockets(t *testing.T) {
	agent := listen(t, "agent.sock")
	docker := listen(t, "docker.sock")
	link := filepath.Join(filepath.Dir(docker), "link.sock")
	if err := os.Symlink(docker, link); err != nil {
		t.Fatal(err)
	}
	fs := Check(context.Background(), allOn, Options{
		Getenv:        env(map[string]string{"SSH_AUTH_SOCK": agent, "DOCKER_HOST": "unix://" + link}),
		DockerSockets: []string{docker, filepath.Join(filepath.Dir(docker), "missing.sock")},
	})
	got := ids(fs)
	if got[SSHAgentSocket] != 1 || got[DockerSocket] != 1 || got[CloudMetadataEndpoint] != 0 {
		t.Errorf("findings = %+v; want one agent and one docker socket (symlink deduplicated), no IMDS", fs)
	}
	for _, f := range fs {
		if f.ID == DockerSocket && f.Severity != rules.SeverityCritical {
			t.Errorf("docker severity = %s", f.Severity)
		}
	}
}

func TestDisabledAndUnreachable(t *testing.T) {
	agent := listen(t, "agent.sock")
	off := allOn
	off.SSHAgent.Enabled = false
	if fs := Check(context.Background(), off, Options{Getenv: env(map[string]string{"SSH_AUTH_SOCK": agent})}); len(fs) != 0 {
		t.Errorf("disabled check reported %+v", fs)
	}
	if fs := Check(context.Background(), allOn, Options{Getenv: env(map[string]string{"SSH_AUTH_SOCK": "/nonexistent/agent.sock"})}); len(fs) != 0 {
		t.Errorf("unreachable socket reported %+v", fs)
	}
	if fs := Check(context.Background(), allOn, Options{}); len(fs) != 0 {
		t.Errorf("nil Getenv: %+v", fs)
	}
}

func TestIMDSProbeIsOptIn(t *testing.T) {
	var requests []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		w.WriteHeader(http.StatusForbidden) // any answer means "reachable"
	}))
	defer srv.Close()

	if fs := Check(context.Background(), allOn, Options{IMDSURL: srv.URL + "/"}); len(fs) != 0 || len(requests) != 0 {
		t.Fatalf("without ProbeIMDS: findings %+v, requests %v", fs, requests)
	}
	fs := Check(context.Background(), allOn, Options{ProbeIMDS: true, IMDSURL: srv.URL + "/"})
	if ids(fs)[CloudMetadataEndpoint] != 1 {
		t.Errorf("findings = %+v", fs)
	}
	if len(requests) != 1 || requests[0] != "GET /" {
		t.Errorf("requests = %v; want exactly one GET of the root", requests)
	}
	if fs := Check(context.Background(), allOn, Options{ProbeIMDS: true, IMDSURL: "http://127.0.0.1:1/"}); len(fs) != 0 {
		t.Errorf("closed port reported %+v", fs)
	}
}
