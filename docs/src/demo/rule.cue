package whac

// A rule for a fictional provider. Each rule has samples.
// whac stops the load if a sample gives the wrong result.
rules: ACME_API_TOKEN: {
	description: "Acme API token"
	severity:    "CRITICAL"
	keywords: ["acme_"]
	regex: #"\b(acme_[A-Za-z0-9]{32})\b"#
	entropy: {min: 3.5, charset: "alnum"}
	tests: {
		mustMatch: ["token = acme_EXAMPLE0123456789abcdefghijklmno"]
		// Matches the regex but fails the entropy check: not a secret.
		mustIgnore: ["token = acme_XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"]
	}
}
