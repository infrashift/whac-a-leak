package whac

// Agent-context preflight (ADR-0011): credential files that an agent's
// process can read. Paths are relative to $HOME and to each --root.
config: credentialPaths: {
	AWS_SSO_CACHE: {glob: ".aws/sso/cache/*.json", description: "AWS SSO token cache"}
	AWS_CLI_CACHE: {glob: ".aws/cli/cache/*.json", description: "AWS CLI credential cache"}
	AWS_CREDENTIALS: {glob: ".aws/credentials", description: "AWS shared credentials file", format: "ini"}
	GCLOUD_ADC: {glob: ".config/gcloud/application_default_credentials.json", description: "gcloud application default credentials"}
	AZURE_MSAL_CACHE: {glob: ".azure/msal_token_cache.json", description: "Azure CLI MSAL token cache"}
	AZURE_ACCESS_TOKENS: {glob: ".azure/accessTokens.json", description: "Azure CLI token cache (older versions)"}
	GH_HOSTS: {glob: ".config/gh/hosts.yml", description: "GitHub CLI credentials"}
	DOCKER_CONFIG: {glob: ".docker/config.json", description: "Docker registry credentials"}
	KUBECONFIG: {glob: ".kube/config", description: "Kubernetes client configuration", format: "yaml"}
	NPMRC: {glob: ".npmrc", description: "npm registry tokens"}
	PYPIRC: {glob: ".pypirc", description: "PyPI upload credentials", format: "ini"}
	NETRC: {glob: ".netrc", description: "netrc machine passwords"}
	GIT_CREDENTIALS: {glob: ".git-credentials", description: "git credential store"}
	TERRAFORM_CREDENTIALS: {glob: ".terraform.d/credentials.tfrc.json", description: "HCP Terraform CLI token"}
	VAULT_TOKEN_FILE: {glob: ".vault-token", description: "Vault CLI token"}
	BASH_HISTORY: {glob: ".bash_history", description: "bash history", maxSize: 16777216}
	ZSH_HISTORY: {glob: ".zsh_history", description: "zsh history", maxSize: 16777216}
	FISH_HISTORY: {glob: ".local/share/fish/fish_history", description: "fish history", maxSize: 16777216}
}

rules: SESSION_TOKEN: {
	description: "Session or OAuth token in a structured credential file"
	kind:        "assignment"
	severity:    "HIGH"
	paths: ["**.accesstoken", "**.access_token", "**.refreshtoken", "**.refresh_token", "**.id_token", "**.oauth_token", "**.client-key-data"]
	entropy: {minLen: 16}
	tests: {
		mustMatch: [
			{file: "sso-cache.json", content: #"{"startUrl": "https://example.awsapps.com/start", "accessToken": "EXAMPLEaB3dE5fG7hJ9kL2mN4pQ6rS8tU"}"#},
			{file: "kubeconfig.yaml", content: "users:\n- name: dev\n  user:\n    client-key-data: RVhBTVBMRS1rZXktZGF0YS1ub3QtcmVhbA==\n"},
		]
		mustIgnore: [
			{file: "sso-cache.json", content: #"{"accessToken": "${ACCESS_TOKEN}"}"#},
			{file: "sso-cache.json", content: #"{"accessToken": "short"}"#},
		]
	}
}

rules: NETRC_PASSWORD: {
	description: "Password in a netrc file"
	severity:    "HIGH"
	keywords: ["machine"]
	multiline: true
	regex: #"(?i)\bmachine\s+\S+(?:\s+(?:login|account)\s+\S+)*\s+password\s+(\S+)"#
	allowRegexes: [#"^\$"#]
	tests: {
		mustMatch: [
			"machine api.example.com login app password EXAMPLE-hunter2",
			"machine api.example.com\n  login app\n  password EXAMPLE-hunter2\n",
		]
		mustIgnore: ["machine learning is not a credential", "machine api.example.com login app password $NETRC_PASSWORD"]
	}
}
