package whac

rules: GENERIC_API_KEY: {
	description: "High-entropy value assigned to an API key, token or secret name"
	severity:    "MEDIUM"
	keywords: ["api", "token", "secret"]
	regex: #"(?i)\b(?:api[_-]?key|api[_-]?secret|access[_-]?token|auth[_-]?token|client[_-]?secret|secret[_-]?key)\b["']?\s*[:=]\s*["']?([A-Za-z0-9+/_\-]{20,})"#
	entropy: {minLen: 20}
	tests: {
		mustMatch: [
			#"api_key = "Zx9EXAMPLEq7Lm2Vt8Rk4Wp6Ys3Hd5Nf""#,
			"CLIENT_SECRET: Zx9EXAMPLEq7Lm2Vt8Rk4Wp6Ys3Hd5Nf",
		]
		mustIgnore: [
			#"api_key = "xxxxxxxxxxxxxxxxxxxxxxxx""#,   // placeholder: low entropy
			#"api_key = "${API_KEY}""#,                  // variable reference
			"access_token: changeme_changeme_changeme", // low entropy
		]
	}
}
