package whac

import "strings"

// Samples are assembled from pieces at load time; see rules_saas.cue.

rules: NPM_TOKEN: {
	description: "npm access token"
	severity:    "CRITICAL"
	keywords: ["npm_"]
	regex: #"\b(npm_[A-Za-z0-9]{36})\b"#
	tests: {
		mustMatch: ["//registry.npmjs.org/:_authToken=" + "np" + "m_" + "EXAMPLE0a1B2c3D4e5F6g7H8i9J0kLmNoPqr"]
		mustIgnore: ["np" + "m_" + "run_build"]
	}
}

rules: PYPI_TOKEN: {
	description: "PyPI upload token"
	severity:    "CRITICAL"
	keywords: ["pypi-ageichlwas5vcmc"]
	regex: #"\b(pypi-AgEIcHlwaS5vcmc[A-Za-z0-9_-]{50,})"#
	tests: {
		mustMatch: ["password = " + "pypi-" + "AgEIcHlwaS5vcmc" + "EXAMPLE" + strings.Repeat("0a1B2c3D4e", 5)]
		mustIgnore: ["pypi-" + "AgEIcHlwaS5vcmc" + "short"]
	}
}

rules: DOCKER_HUB_TOKEN: {
	description: "Docker Hub personal access token"
	severity:    "HIGH"
	keywords: ["dckr_pat_"]
	regex: #"\b(dckr_pat_[A-Za-z0-9_-]{27})(?:[^A-Za-z0-9_-]|$)"#
	tests: {
		mustMatch: ["DOCKER_TOKEN=" + "dckr_" + "pat_" + "EXAMPLE0a1B2c3D4e5F6g7H8i9J"]
		mustIgnore: ["dckr_" + "pat_" + "short"]
	}
}

rules: DOCKER_CONFIG_AUTH: {
	description: "Registry credentials in a Docker config file"
	severity:    "HIGH"
	keywords: [#""auth""#]
	regex: #""auth"\s*:\s*"([A-Za-z0-9+/]{16,}={0,2})""#
	tests: {
		// base64("example:EXAMPLE-password")
		mustMatch: [{file: "config.json", content: #"{"auths": {"registry.example.com": {"auth": "ZXhhbXBsZTpFWEFNUExFLXBhc3N3b3Jk"}}}"#}]
		mustIgnore: [{file: "config.json", content: #"{"auths": {"registry.example.com": {"auth": ""}}}"#}]
	}
}
