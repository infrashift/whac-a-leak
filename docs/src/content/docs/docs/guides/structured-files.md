---
title: Structured Files
description: How whac reads YAML, JSON, .env, properties, INI and HCL files as key paths.
---

A line regex loses the relation between a key and its value. whac parses
structured files into key paths. An `assignment` rule then checks each value.
This finds a plain password at `database.credentials.password`, even with low
entropy and no token prefix.

## Formats

| Format | File names | Key path example |
|---|---|---|
| YAML | `*.yaml`, `*.yml` | `database.credentials.password` |
| JSON | `*.json` | `oauth.refreshToken` |
| dotenv | `.env`, `.env.*`, `*.env` | `DB_PASSWORD` |
| Java properties | `*.properties` | `db.password` |
| INI | `*.ini`, `*.cfg` | `database.password` (section, then key) |
| HCL | `*.tf`, `*.tfvars`, `*.hcl` | `resource.aws_db_instance.main.password` |

Sequence elements use their index, for example `servers.0.token`. whac reads
quoted HCL strings only. It skips unquoted HCL values, because they are
references or expressions.

If a structured file does not parse, whac treats it as unstructured text.

## Unstructured text

For other files, whac extracts `key = value`, `key: value`, `key := value`
and `key => value` pairs. It does this only on lines that hold a keyword of an
assignment rule. It skips comparisons such as `==` and `!=`.

## Builtin assignment rules

| Rule | Selects | Value check |
|---|---|---|
| `SENSITIVE_PATH` | Key paths in `config.sensitivePaths` | Placeholder filter only. Low entropy is enough. |
| `SECRET_ASSIGNMENT` | Keys that look like `secret`, `token`, `api_key`, `password` | Placeholder filter, safe formats, entropy. |

`SENSITIVE_PATH` has no keywords. It therefore runs on structured files only.

## Placeholder filter

An assignment rule drops these values:

* Empty values.
* Variable references: `${VAR}`, `$VAR`, `%VAR%`, `{{ .Value }}`.
* Angle placeholders: `<your-token>`.
* Placeholder words: `changeme`, `password`, `secret`, `todo`, `null` and similar.
* Placeholder text: values that start with `your-`, `insert-`, `enter-` or `replace-`.
* Masked values: values that contain `xxxx` or `****`.
* Values that repeat one character.

## Safe formats

A rule with an `entropy` block also drops these formats:

* UUIDs.
* File paths, such as `/var/run/secrets/token`.
* URLs without credentials. A URL with `user:password@` stays a finding.
* Hex hashes (32 to 128 digits) under a key that contains `sha`, `md5`,
  `hash`, `digest`, `checksum`, `integrity`, `etag`, `commit` or `rev`.

## Add a sensitive path

1. Create the file `.whac/paths.cue`.
2. Add this layer:

   ```cue
   package whac

   config: sensitivePaths: "**.vault.unseal_key": "Vault unseal keys"
   ```

3. Run `whac scan`.

* **Expected result:** whac reports each value at `vault.unseal_key` with the
  rule `SENSITIVE_PATH`.

A key-path glob uses `.` between segments. `*` matches part of one segment.
`**` matches any number of segments. The match ignores case.

## One finding per span

A pattern rule and an assignment rule can report the same value. whac keeps
one finding per span: the higher severity, then the lower rule ID.
