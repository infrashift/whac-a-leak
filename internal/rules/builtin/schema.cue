// Whac-a-Leak rule and configuration schema.
//
// This file and the rule files next to it are embedded in the whac binary.
// User layers in .whac/*.cue (package whac) are unified with them at load
// time; see .okfw/specs/rule-schema for the semantics. internal/rules/spec.go
// mirrors these definitions for decoding.
package whac

#Severity: "LOW" | "MEDIUM" | "HIGH" | "CRITICAL"
#Charset:  "hex" | "alnum" | "base64" | "base64url"

#SecretRule: {
	id:          =~"^[A-Z][A-Z0-9_]*$"
	description: string & !=""
	kind:        *"pattern" | "assignment"
	severity:    *"HIGH" | #Severity
	enabled:     *true | bool

	// Literal keywords for the Aho-Corasick prefilter, matched
	// case-insensitively. Pattern rules need at least one.
	keywords: [...(string & !="")]
	// RE2 syntax (Go regexp). Required for pattern rules.
	regex?: string
	// Capture group holding the secret; 0 means the whole match.
	secretGroup: *1 | (int & >=0)
	// The engine evaluates the regex on each line that holds a keyword. Set
	// multiline for secrets that span lines, such as PEM keys: the window then
	// runs up to 64 KiB from the start of the keyword's line.
	multiline: *false | bool
	// Assignment rules (Milestone 3): a regex on the key name and/or
	// doublestar globs over the dotted key path.
	keyContext?: string
	paths?: [...string]

	entropy?: {
		// Bits per character. When min is omitted, the engine derives a
		// length- and charset-aware default (ADR-0006).
		min?:     number & >=0 & <=8
		max?:     number & >=0 & <=8
		minLen:   *16 | (int & >0)
		charset?: #Charset
		if min != _|_ if max != _|_ {max: >=min}
	}
	// Case-insensitive substrings that, when present in the secret, drop it.
	stopwords: [...(string & !="")]
	// Secrets matching any of these regexes are dropped.
	allowRegexes: [...string]

	tests: {
		// Each sample must produce a finding. Use obviously fake values.
		mustMatch: [string, ...string]
		// Each sample must produce no finding (often: matches the regex but
		// is rejected by entropy, stopwords or allowRegexes).
		mustIgnore: [...string]
	}

	if kind == "pattern" {
		keywords: [string, ...string]
		regex:    string
		// Cheap author-side check under `cue vet`; the authoritative check
		// runs every sample through the real engine (ADR-0009).
		tests: mustMatch: [...=~regex]
	}
}

// Re-grading or disabling a rule from a layer. A rule's own severity is
// concrete, so it cannot be changed by unification; overrides can.
#Override: {
	severity?: #Severity
	enabled?:  bool
}

#Config: {
	// Files larger than this many bytes are skipped.
	maxFileSize: *2097152 | (int & >0)
	// Doublestar globs → reason.
	exclude: [Glob=string]: string & !=""
}

#Allowlist: {
	// whac1 fingerprint → reason. Only for false positives: real secrets
	// get rotated, never allowlisted (ADR-0007).
	fingerprints: [=~"^whac1:[0-9a-f]{64}$"]: string & !=""
	// Doublestar path globs → reason.
	paths: [Glob=string]: string & !=""
	// Regexes over the secret → reason.
	regexes: [Re=string]: string & !=""
}

rules: [ID=string]: #SecretRule & {id: ID}
overrides: [ID=string]: #Override
config:    #Config
allowlist: #Allowlist
