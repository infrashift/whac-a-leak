package whac

rules: GITHUB_TOKEN: {
	description: "GitHub personal access, OAuth, app or refresh token"
	severity:    "CRITICAL"
	keywords: ["ghp_", "gho_", "ghu_", "ghs_", "ghr_", "github_pat_"]
	regex: #"\b((?:ghp|gho|ghu|ghs|ghr)_[A-Za-z0-9]{36}|github_pat_[A-Za-z0-9]{22}_[A-Za-z0-9]{59})\b"#
	entropy: {minLen: 40}
	tests: {
		// Fake: the checksum GitHub embeds in real tokens is wrong.
		mustMatch: ["GH_TOKEN=ghp_EXAMPLE0a1B2c3D4e5F6g7H8i9J0kLmNoPqr"]
		mustIgnore: [
			"GH_TOKEN=ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", // placeholder
			"GH_TOKEN=ghp_tooShort",
		]
	}
}
