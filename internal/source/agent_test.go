package source

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/infrashift/whac-a-leak/internal/rules"
)

func TestEnvironment(t *testing.T) {
	content, names := Environment([]string{"Z=last", "A=first", "MULTI=line1\nline2", "=bad", "NOEQUALS"})
	if string(content) != "A=first\nMULTI=line1\nline2\nZ=last\n" {
		t.Errorf("content = %q", content)
	}
	if !slices.Equal(names, []string{"A", "MULTI", "MULTI", "Z"}) {
		t.Errorf("names = %v", names)
	}
}

func TestCredentialFiles(t *testing.T) {
	home := t.TempDir()
	mount := t.TempDir()
	writeFile(t, home, ".aws/credentials", "x")
	writeFile(t, home, ".aws/sso/cache/a.json", "{}")
	writeFile(t, home, ".aws/sso/cache/b.json", "{}")
	writeFile(t, mount, ".netrc", "x")
	if err := os.MkdirAll(filepath.Join(home, ".kube", "config"), 0o750); err != nil { // a directory, not a file
		t.Fatal(err)
	}
	paths := map[string]rules.CredentialPath{
		"AWS":     {Glob: ".aws/credentials", Enabled: true, Format: "ini", MaxSize: 10},
		"SSO":     {Glob: ".aws/sso/cache/*.json", Enabled: true},
		"NETRC":   {Glob: ".netrc", Enabled: true},
		"KUBE":    {Glob: ".kube/config", Enabled: true},
		"OFF":     {Glob: ".aws/credentials", Enabled: false},
		"WINDOWS": {Glob: ".netrc", Enabled: true, OS: "windows"},
	}
	files, err := CredentialFiles([]string{home, mount, filepath.Join(home, "missing")}, home, paths, "linux")
	if err != nil {
		t.Fatal(err)
	}
	got := make([]string, 0, len(files))
	for _, f := range files {
		got = append(got, f.ID+" "+f.Path)
	}
	want := []string{"AWS ~/.aws/credentials", "SSO ~/.aws/sso/cache/a.json", "SSO ~/.aws/sso/cache/b.json", "NETRC " + filepath.ToSlash(filepath.Join(mount, ".netrc"))}
	if !slices.Equal(got, want) {
		t.Errorf("files =\n%v\nwant\n%v", got, want)
	}
	if files[0].Format != "ini" || files[0].MaxSize != 10 || files[0].Rel != ".aws/credentials" {
		t.Errorf("first = %+v", files[0])
	}
}
