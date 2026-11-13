package report

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	cueerrors "cuelang.org/go/cue/errors"
	cuejson "cuelang.org/go/encoding/json"
	"cuelang.org/go/encoding/jsonschema"

	"github.com/infrashift/whac-a-leak/internal/rules"
	"github.com/infrashift/whac-a-leak/internal/scan"
)

// sarifSchemaValue compiles the official SARIF 2.1.0 JSON Schema (OASIS,
// vendored in testdata/sarif) with CUE's JSON Schema decoder.
func sarifSchemaValue(t *testing.T) (*cue.Context, cue.Value) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "sarif", "sarif-schema-2.1.0.json"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := cuecontext.New()
	expr, err := cuejson.Extract("sarif-schema-2.1.0.json", data)
	if err != nil {
		t.Fatal(err)
	}
	f, err := jsonschema.Extract(ctx.BuildExpr(expr), &jsonschema.Config{})
	if err != nil {
		t.Fatalf("extract schema: %v", cueerrors.Details(err, nil))
	}
	schema := ctx.BuildFile(f)
	if err := schema.Err(); err != nil {
		t.Fatalf("build schema: %v", cueerrors.Details(err, nil))
	}
	return ctx, schema
}

func TestSARIFValidatesAgainstSchema(t *testing.T) {
	res := result()
	res.Findings = append(res.Findings, scan.Finding{
		RuleID: "SENSITIVE_PATH", Description: "Value at a sensitive key path", Severity: rules.SeverityHigh,
		Path: "config.yaml", Line: 3, Column: 15, EndLine: 3, Fingerprint: "whac1:cd", KeyPath: "db.password",
		Redacted: "****", Commit: strings.Repeat("a", 40), Author: "Ada", Date: "2026-01-01T00:00:00Z",
	})
	var b bytes.Buffer
	opts := Options{NoRedact: true, ToolVersion: "v1.2.3", Rules: []*rules.Rule{{ID: "GITHUB_TOKEN", Description: "GitHub token", Severity: rules.SeverityCritical}}}
	if err := SARIF(&b, res, opts); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(b.String(), "SECRETVALUE") {
		t.Error("SARIF must redact secrets even with NoRedact")
	}

	ctx, schema := sarifSchemaValue(t)
	doc := ctx.CompileBytes(b.Bytes())
	if err := schema.Unify(doc).Validate(cue.Concrete(true)); err != nil {
		t.Fatalf("SARIF output does not match the schema:\n%s\n%s", cueerrors.Details(err, nil), b.String())
	}

	var log struct {
		Runs []struct {
			Tool struct {
				Driver struct{ Rules []struct{ ID string } }
			}
			Results []struct {
				RuleIndex           int
				Level               string
				PartialFingerprints map[string]string
				Properties          map[string]string
			}
		}
	}
	if err := json.Unmarshal(b.Bytes(), &log); err != nil {
		t.Fatal(err)
	}
	run := log.Runs[0]
	if len(run.Tool.Driver.Rules) != 2 || run.Tool.Driver.Rules[1].ID != "SENSITIVE_PATH" {
		t.Errorf("rules = %+v; a finding's unknown rule must be added", run.Tool.Driver.Rules)
	}
	if r := run.Results[1]; r.RuleIndex != 1 || r.Level != "error" || r.PartialFingerprints[fingerprintKey] != "whac1:cd" || r.Properties["author"] != "Ada" {
		t.Errorf("result = %+v", r)
	}
}

func TestSARIFRejectsBrokenDocument(t *testing.T) {
	// Guards the validation itself: a document without "version" must fail.
	ctx, schema := sarifSchemaValue(t)
	doc := ctx.CompileString(`{"runs": []}`)
	if err := schema.Unify(doc).Validate(cue.Concrete(true)); err == nil {
		t.Error("the schema accepted a SARIF log without a version")
	}
}

func TestSARIFLevels(t *testing.T) {
	want := map[rules.Severity]string{rules.SeverityCritical: "error", rules.SeverityHigh: "error", rules.SeverityMedium: "warning", rules.SeverityLow: "note"}
	for s, l := range want {
		if sarifLevel(s) != l || securitySeverity(s) == "" {
			t.Errorf("severity %s: level %s", s, sarifLevel(s))
		}
	}
}
