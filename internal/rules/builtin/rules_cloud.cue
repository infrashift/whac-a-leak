package whac

rules: AWS_ACCESS_KEY_ID: {
	description: "AWS access key ID"
	severity:    "CRITICAL"
	keywords: ["akia", "asia", "abia", "acca"]
	regex: #"\b((?:AKIA|ASIA|ABIA|ACCA)[A-Z2-7]{16})\b"#
	entropy: {minLen: 20, charset: "alnum"}
	tests: {
		// AWS's own documentation example key.
		mustMatch: ["aws_access_key_id = AKIAIOSFODNN7EXAMPLE"]
		mustIgnore: [
			"aws_access_key_id = AKIAXXXXXXXXXXXXXXXX", // placeholder: too little entropy
			"aws_access_key_id = AKIAIOSFODNN7EXAMPL",  // too short
		]
	}
}
