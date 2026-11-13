---
title: Agent-Context Preflight
description: Before an AI agent starts, check that its environment holds no credentials it does not need.
---

A coding agent can read what its process can reach: environment variables,
files of its user, mounted volumes and credential sockets.
`whac scan --agent-context` audits that context. It exits with code `1` if
anything is reachable.

whac audits the context. The environment builder makes the context clean:
an allowlist of environment variables (`env -i`), scoped mounts, and
short-lived credentials with a narrow scope.

## What whac checks

| Check | Flag | Finding |
|---|---|---|
| Environment variables | `--env` | A secret in a variable. The key path is `env.<NAME>`. |
| Credential files under `$HOME` and each `--root` | `--home` | A secret in a known credential file. |
| SSH agent socket (`SSH_AUTH_SOCK`) | `--agent-context` | The process can sign with the loaded SSH keys. |
| Docker socket | `--agent-context` | The process controls the Docker daemon, which is root on the host. |
| Cloud metadata endpoint (IMDS) | `--agent-context --probe-imds` | The process can request instance credentials. |

`--agent-context` runs all checks except the IMDS probe. The IMDS probe is
the only network request that whac makes. It sends one `GET` to the
endpoint root and requests no credential path.

whac never reads the memory of another process.

## Values are always masked

Preflight output often goes to logs that the agent can read. whac therefore
prints `[masked]` for every value, in every format. `--no-redact` has no
effect in this mode. Each finding names the variable, the file or the socket,
the rule and the fingerprint.

## Credential files

whac checks the paths in `config.credentialPaths`, relative to `$HOME` and to
each `--root`. The builtin list covers these files:

* AWS: `.aws/credentials`, `.aws/sso/cache/*.json`, `.aws/cli/cache/*.json`
* Google Cloud: `.config/gcloud/application_default_credentials.json`
* Azure: `.azure/msal_token_cache.json`, `.azure/accessTokens.json`
* GitHub CLI, Docker, Kubernetes: `.config/gh/hosts.yml`, `.docker/config.json`, `.kube/config`
* Package registries: `.npmrc`, `.pypirc`
* Other stores: `.netrc`, `.git-credentials`, `.vault-token`, `.terraform.d/credentials.tfrc.json`
* Shell history: `.bash_history`, `.zsh_history`, `.local/share/fish/fish_history`

To add a file, add a layer:

```cue
package whac

config: credentialPaths: ACME_CLI: {
	glob:        ".config/acme/credentials.json"
	description: "Acme CLI login"
}
```

To skip a builtin entry, set `config: credentialPaths: BASH_HISTORY: enabled: false`.
To skip an exposure check, set `config: exposure: dockerSocket: enabled: false`.

## Run the preflight in a devcontainer

1. Add the command to `.devcontainer/devcontainer.json`:

   ```json
   {
     "postStartCommand": "whac scan --agent-context"
   }
   ```

2. Start the container.

* **Expected result:** the start fails when whac finds a credential, and the
  log names it.

## Run the preflight in a microVM

1. Add this step to the init script, before the agent starts:

   ```sh
   whac scan --agent-context --root /mnt/workspace --probe-imds || exit 1
   ```

2. Boot the microVM.

* **Expected result:** the init script stops when whac finds a credential.
  The agent does not start.

Run whac as the agent's user, with the agent's environment and mounts. A
check as another user audits another context.

## Accept a finding

Some exposures are on purpose, for example a forwarded SSH agent. Allowlist
the finding by its fingerprint:

```sh
whac allow add whac1:<fingerprint> --reason "agent forwarding for git push"
```
