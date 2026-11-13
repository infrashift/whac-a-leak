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

// Samples for providers that GitHub push protection detects are assembled
// from pieces at load time, so no complete token appears in this file.

rules: AWS_SECRET_ACCESS_KEY: {
	description: "AWS secret access key"
	severity:    "CRITICAL"
	keywords: ["aws_secret", "secretaccesskey", "secret_access_key"]
	regex: #"(?i)\baws_?secret_?(?:access_?)?key["']?\s*[:=]\s*["']?([A-Za-z0-9/+=]{40})(?:[^A-Za-z0-9/+=]|$)"#
	entropy: {minLen: 40}
	tests: {
		mustMatch: ["aws_secret_access_key = " + "EXAMPLE" + "kQ7v/Lm2+Rt9Xw4Zp1Nb8Hc3Jd6Fg0Ys5"]
		mustIgnore: [
			"aws_secret_access_key = ${AWS_SECRET_ACCESS_KEY}",
			"aws_secret_access_key = EXAMPLEEXAMPLEEXAMPLEEXAMPLEEXAMPLEEXAMP", // low entropy
		]
	}
}

rules: GCP_API_KEY: {
	description: "Google Cloud API key"
	severity:    "HIGH"
	keywords: ["aiza"]
	regex: #"\b(AIza[0-9A-Za-z_-]{35})(?:[^0-9A-Za-z_-]|$)"#
	tests: {
		mustMatch: ["key=" + "AI" + "za" + "SyEXAMPLE0a1B2c3D4e5F6g7H8i9J0kLmNo"]
		mustIgnore: ["key=" + "AI" + "zaSyShort"]
	}
}

rules: GCP_SERVICE_ACCOUNT_KEY: {
	description: "Google Cloud service account private key in JSON"
	severity:    "CRITICAL"
	keywords: [#""private_key""#]
	regex: #""private_key"\s*:\s*"(-----BEGIN PRIVATE KEY-----\\n(?:[A-Za-z0-9+/=]+\\n)+-----END PRIVATE KEY-----\\n?)""#
	tests: {
		mustMatch: [{file: "service-account.json", content: #"""
			{
			  "type": "service_account",
			  "private_key": "-----BEGIN PRIVATE KEY-----\nEXAMPLEKEYbodyNotARealKey0123456789abcdefghijklmn\nEXAMPLEKEYbodyNotARealKey0123456789opqrstuvwxyzAB\n-----END PRIVATE KEY-----\n"
			}
			"""#}]
		mustIgnore: [{file: "service-account.json", content: #"{"private_key": ""}"#}]
	}
}

rules: AZURE_STORAGE_ACCOUNT_KEY: {
	description: "Azure storage account key in a connection string"
	severity:    "CRITICAL"
	keywords: ["accountkey="]
	regex: #"(?i)AccountKey=([A-Za-z0-9+/]{86}==)"#
	tests: {
		mustMatch: ["DefaultEndpointsProtocol=https;AccountName=example;Account" + "Key=" +
			"EXAMPLE0a1B2c3D4e5F6g7H8i9J0kLmNoPqrStUvWxYz" + "Q9wv7Lm2Rt5Kp8Zx3Nb6Hc1Jd4Fg7Ys0Te2Ua5WAbc==" +
			";EndpointSuffix=core.windows.net"]
		mustIgnore: ["AccountKey=<storage-account-key>"]
	}
}

rules: AZURE_CLIENT_SECRET: {
	description: "Microsoft Entra ID (Azure AD) client secret"
	severity:    "HIGH"
	keywords: ["q~"]
	regex: #"(?:^|[^A-Za-z0-9_~.-])([A-Za-z0-9_~.]{3}[0-9]Q~[A-Za-z0-9_~.-]{31,34})(?:[^A-Za-z0-9_~.-]|$)"#
	tests: {
		mustMatch: ["AZURE_CLIENT_SECRET=" + "abc8" + "Q~" + "EXAMPLE0a1B2c3D4e5F6g7H8i9J0kLmN"]
		mustIgnore: ["home dir: ~/Q~notasecret"]
	}
}

rules: HEROKU_API_KEY: {
	description: "Heroku API key"
	severity:    "HIGH"
	keywords: ["heroku"]
	// A UUID: the rule has no entropy block, so the UUID safe format does not apply.
	regex: #"(?i)heroku[^\n]{0,40}?\b([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})\b"#
	tests: {
		mustMatch: ["HEROKU_API_KEY=01234567-89ab-cdef-0123-456789abcdef"]
		mustIgnore: ["heroku app: whac-demo"]
	}
}
