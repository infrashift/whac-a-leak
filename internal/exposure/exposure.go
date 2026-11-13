// Package exposure checks for reachable access to credentials: sockets and
// endpoints that hand out credentials without a secret string (ADR-0011).
package exposure

import (
	"context"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/infrashift/whac-a-leak/internal/rules"
)

// Check IDs.
const (
	SSHAgentSocket        = "SSH_AGENT_SOCKET"
	DockerSocket          = "DOCKER_SOCKET"
	CloudMetadataEndpoint = "CLOUD_METADATA_ENDPOINT"
)

// DefaultIMDSURL is the link-local cloud metadata endpoint (AWS, GCP, Azure).
const DefaultIMDSURL = "http://169.254.169.254/"

// DefaultDockerSockets are the usual Docker daemon socket paths.
var DefaultDockerSockets = []string{"/var/run/docker.sock", "/run/docker.sock"}

// Finding is one reachable access path to credentials.
type Finding struct {
	ID          string
	Description string
	Severity    rules.Severity
	// Path is the socket path or the endpoint URL.
	Path string
}

// Options controls the checks. Tests replace the environment, socket paths
// and endpoint.
type Options struct {
	Getenv        func(string) string
	DockerSockets []string
	// ProbeIMDS enables the one network request: a GET to IMDSURL.
	ProbeIMDS bool
	IMDSURL   string
	Timeout   time.Duration
}

// Check runs the enabled checks.
func Check(ctx context.Context, cfg rules.Exposure, o Options) []Finding {
	if o.Getenv == nil {
		o.Getenv = func(string) string { return "" }
	}
	if o.Timeout == 0 {
		o.Timeout = 500 * time.Millisecond
	}
	var out []Finding
	if cfg.SSHAgent.Enabled {
		if p := o.Getenv("SSH_AUTH_SOCK"); p != "" && connectable(ctx, p, o.Timeout) {
			out = append(out, Finding{
				ID: SSHAgentSocket, Severity: cfg.SSHAgent.Severity, Path: p,
				Description: "SSH agent socket is reachable: the process can sign with the loaded SSH keys",
			})
		}
	}
	if cfg.DockerSocket.Enabled {
		sockets := slices.Clone(o.DockerSockets)
		if h := o.Getenv("DOCKER_HOST"); strings.HasPrefix(h, "unix://") {
			sockets = append(sockets, strings.TrimPrefix(h, "unix://"))
		}
		// /var/run is often a symlink to /run: report each socket once.
		for i, p := range sockets {
			if r, err := filepath.EvalSymlinks(p); err == nil {
				sockets[i] = r
			}
		}
		slices.Sort(sockets)
		for _, p := range slices.Compact(sockets) {
			if connectable(ctx, p, o.Timeout) {
				out = append(out, Finding{
					ID: DockerSocket, Severity: cfg.DockerSocket.Severity, Path: p,
					Description: "Docker socket is reachable: the process controls the Docker daemon, which is root on the host",
				})
			}
		}
	}
	if cfg.CloudMetadata.Enabled && o.ProbeIMDS {
		url := o.IMDSURL
		if url == "" {
			url = DefaultIMDSURL
		}
		if reachable(ctx, url, o.Timeout) {
			out = append(out, Finding{
				ID: CloudMetadataEndpoint, Severity: cfg.CloudMetadata.Severity, Path: url,
				Description: "Cloud metadata endpoint is reachable: the process can request instance credentials",
			})
		}
	}
	return out
}

// connectable reports whether a Unix socket accepts a connection.
func connectable(ctx context.Context, path string, timeout time.Duration) bool {
	d := net.Dialer{Timeout: timeout}
	c, err := d.DialContext(ctx, "unix", path)
	if err != nil {
		return false
	}
	_ = c.Close()
	return true
}

// reachable sends one GET to the endpoint root and reports whether any HTTP
// response came back. It requests no credential path, sends no secret and
// follows no redirect.
func reachable(ctx context.Context, url string, timeout time.Duration) bool {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return false
	}
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
	_ = resp.Body.Close()
	return true
}
