package whac

import "strings"

// Samples are assembled from pieces at load time; see rules_saas.cue.

rules: VAULT_TOKEN: {
	description: "HashiCorp Vault service, batch or recovery token"
	severity:    "CRITICAL"
	keywords: ["hvs.", "hvb.", "hvr."]
	regex: #"\b(hv[sbr]\.[A-Za-z0-9_-]{24,})"#
	tests: {
		mustMatch: ["VAULT_TOKEN=" + "hv" + "s." + "EXAMPLE0a1B2c3D4e5F6g7H8i9J0kLmN"]
		mustIgnore: ["hv" + "s.short"]
	}
}

rules: TERRAFORM_CLOUD_TOKEN: {
	description: "HCP Terraform (Terraform Cloud) API token"
	severity:    "CRITICAL"
	keywords: [".atlasv1."]
	regex: #"\b([A-Za-z0-9]{14}\.atlasv1\.[A-Za-z0-9_=-]{60,70})"#
	tests: {
		mustMatch: ["token = \"" + "EXAMPLE0a1B2c3" + ".atlas" + "v1." + "EXAMPLE0a1B2c3D4e5F6g7H8i9J0kLmNoPqrStUvWxYzQ9wv7Lm2Rt5Kp8Zx3Nb" + "\""]
		mustIgnore: ["see .atlas" + "v1. tokens in the docs"]
	}
}

rules: AGE_SECRET_KEY: {
	description: "age encryption secret key"
	severity:    "CRITICAL"
	keywords: ["age-secret-key-1"]
	regex: #"\b(AGE-SECRET-KEY-1[QPZRY9X8GF2TVDW0S3JN54KHCE6MUA7L]{58})\b"#
	tests: {
		mustMatch: ["AGE-" + "SECRET-KEY-1" + "QPZRY9X8GF2TVDW0S3JN54KHCE6MUA7L" + "QPZRY9X8GF2TVDW0S3JN54KHCE"]
		mustIgnore: ["AGE-" + "SECRET-KEY-1" + strings.Repeat("Q", 10)]
	}
}

rules: URL_CREDENTIALS: {
	description: "Password in a URL (user:password@host)"
	severity:    "HIGH"
	keywords: ["://"]
	regex: #"\b[A-Za-z][A-Za-z0-9+.-]*://[^\s:@/]+:([^\s:@/]{3,})@[A-Za-z0-9.-]+"#
	allowRegexes: [#"^\$\{"#, #"^\$[A-Za-z_]"#, #"^<[^>]*>$"#, #"^%[0-9A-Fa-f]{2}"#, #"(?i)^(password|passwd|pass|pwd|secret|changeme|x+|\*+)$"#]
	tests: {
		mustMatch: ["DATABASE_URL=postgres://app:EXAMPLE-hunter2@db.internal:5432/app"]
		mustIgnore: [
			"DATABASE_URL=postgres://app:${DB_PASSWORD}@db.internal/app",
			"https://user:password@example.com/",
			"https://example.com/path:with@colon",
		]
	}
}
