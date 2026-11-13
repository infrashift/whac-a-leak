package cli

import (
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fakeHome(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	p := filepath.Join(home, ".aws", "credentials")
	if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("[default]\naws_access_key_id = AKIAIOSFODNN7EXAMPLE\n"), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestScanHome(t *testing.T) {
	fakeHome(t)
	mount := t.TempDir()
	if err := os.WriteFile(filepath.Join(mount, ".git-credentials"), []byte("https://app:EXAMPLE-hunter2@git.example.com\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	code, out, errOut := run(t, "scan", "--home", "--root", mount, "--no-redact", "--config-dir", t.TempDir())
	if code != ExitFindings {
		t.Fatalf("code %d, stderr %q", code, errOut)
	}
	for _, want := range []string{"AWS_ACCESS_KEY_ID  ~/.aws/credentials:2:21", "URL_CREDENTIALS", "secret: [masked]"} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "AKIAIOSFODNN7EXAMPLE") || strings.Contains(out, "hunter2") {
		t.Errorf("--no-redact must not print values in agent-context mode:\n%s", out)
	}
}

func TestScanAgentContextSSHAgent(t *testing.T) {
	fakeHome(t)
	dir, err := os.MkdirTemp("", "whac") //nolint:usetesting // short path for the Unix socket limit
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	sock := filepath.Join(dir, "agent.sock")
	var lc net.ListenConfig
	l, err := lc.Listen(t.Context(), "unix", sock)
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
	t.Setenv("SSH_AUTH_SOCK", sock)
	// The real Docker socket of the machine is out of the test's control.
	layers := layerDir(t, map[string]string{"org.cue": "package whac\nconfig: exposure: dockerSocket: enabled: false\n"})
	code, out, errOut := run(t, "scan", "--agent-context", "--format", "json", "--config-dir", layers)
	if code != ExitFindings || !strings.Contains(out, `"ruleId": "SSH_AGENT_SOCKET"`) || !strings.Contains(out, `"category": "exposure"`) {
		t.Errorf("code %d, stderr %q, out:\n%s", code, errOut, out)
	}
	if strings.Contains(out, "DOCKER_SOCKET") {
		t.Error("a disabled exposure check ran")
	}
}

func TestScanAgentFlagErrors(t *testing.T) {
	tests := [][]string{
		{"scan", "--agent-context", "--staged"},
		{"scan", "--home", "--git"},
		{"scan", "--agent-context", "somepath"},
		{"scan", "--probe-imds"},
		{"scan", "--root", "/tmp"},
		{"scan", "--home", "--root", "/nonexistent-root"},
	}
	for _, args := range tests {
		if code, _, _ := run(t, args...); code != ExitError {
			t.Errorf("%v: code %d, want %d", args, code, ExitError)
		}
	}
}
