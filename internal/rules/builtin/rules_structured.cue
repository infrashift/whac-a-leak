package whac

config: sensitivePaths: {
	"**.*password": "password values"
	"**.*passwd":   "password values"
}

rules: SENSITIVE_PATH: {
	description: "Value at a sensitive key path in a structured file"
	kind:        "assignment"
	severity:    "HIGH"
	// Layers extend config.sensitivePaths; this rule follows.
	paths: [for glob, _ in config.sensitivePaths {glob}]
	tests: {
		mustMatch: [
			{file: "config.yaml", content: """
				database:
				  credentials:
				    password: EXAMPLE-hunter2
				"""},
			{file: "app.env", content: "DB_PASSWORD=EXAMPLE-letmein\n"},
			{file: "main.tf", content: """
				resource "db" "main" {
				  password = "EXAMPLE-hunter2"
				}
				"""},
		]
		mustIgnore: [
			{file: "config.yaml", content: "database:\n  password: ${DB_PASSWORD}\n"},
			{file: "config.yaml", content: "password: changeme\n"},
			{file: "config.yaml", content: "username: EXAMPLE-admin\n"},
			{file: "main.tf", content: "password = var.db_password\n"},
			// Path rules need structure: plain text has no key path.
			"password: EXAMPLE-hunter2",
		]
	}
}

rules: SECRET_ASSIGNMENT: {
	description: "High-entropy value assigned to a secret-like key"
	kind:        "assignment"
	severity:    "MEDIUM"
	keywords: ["secret", "token", "key", "passw", "pwd", "auth", "credential"]
	keyContext: #"(?i)(secret|token|api[_-]?key|access[_-]?key|private[_-]?key|passw(or)?d|pwd|auth|credential)"#
	entropy: {minLen: 16}
	tests: {
		mustMatch: [
			{file: "app.yaml", content: "service:\n  auth_token: Zx9EXAMPLEq7Lm2Vt8Rk4Wp6\n"},
			{file: "app.json", content: #"{"oauth": {"clientSecret": "Zx9EXAMPLEq7Lm2Vt8Rk4Wp6"}}"#},
			#"const sessionSecret = "Zx9EXAMPLEq7Lm2Vt8Rk4Wp6Ys3""#,
		]
		mustIgnore: [
			#"auth_token = "550e8400-e29b-41d4-a716-446655440000""#,                       // UUID
			#"auth_url = "https://auth.example.com/oauth/token/refresh/endpoint""#,      // URL without credentials
			#"token_file = "/var/run/secrets/kubernetes.io/serviceaccount/token""#,      // file path
			"token_sha256: 9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08", // hash under a hash key
			"auth_token: <your-token-here-please>",                                        // placeholder
			#"api_key = "${API_KEY_FROM_ENVIRONMENT}""#,                                   // variable reference
			#"secret_key = "aaaaaaaaaaaaaaaaaaaaaaaa""#,                                   // low entropy
			#"if token == "Zx9EXAMPLEq7Lm2Vt8Rk4Wp6Ys3" {"#,                               // comparison, not an assignment
		]
	}
}
