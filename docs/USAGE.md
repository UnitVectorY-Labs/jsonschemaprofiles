---
layout: default
title: Usage
nav_order: 3
permalink: /usage
---

# Usage
{: .no_toc }

## Table of Contents
{: .no_toc .text-delta }

- TOC
{:toc}

---

## Commands

### `profiles list`

List all available profile IDs and their descriptions.

```bash
jsonschemaprofiles profiles list
```

### `validate schema`

Validate a candidate JSON Schema against a profile (Phase 1 + Phase 2).

```bash
jsonschemaprofiles validate schema --profile OPENAI_202602 --in schema.json
```

| Flag | Description |
|---|---|
| `--profile` | Profile ID (required) |
| `--in` | Input file path, or `-` for stdin |
| `--format` | Output format: `text` (default) or `json` |
| `--strict` | Treat warnings as errors |
| `--model-target` | Apply model-specific restrictions |

**Examples**

```bash
# Validate from file
jsonschemaprofiles validate schema --profile OPENAI_202602 --in schema.json

# Validate from stdin
cat schema.json | jsonschemaprofiles validate schema --profile GEMINI_202602 --in -

# JSON output
jsonschemaprofiles validate schema --profile OPENAI_202602 --in schema.json --format json

# Strict mode
jsonschemaprofiles validate schema --profile OPENAI_202602 --in schema.json --strict
```

### `coerce schema`

Produce a coerced schema and a report of changes.

```bash
jsonschemaprofiles coerce schema --profile OPENAI_202602 --in schema.json --out coerced.json
```

| Flag | Description |
|---|---|
| `--profile` | Profile ID (required) |
| `--in` | Input file path, or `-` for stdin |
| `--out` | Output file for coerced schema (default: stdout) |
| `--dry-run` | Show proposed changes without applying |
| `--mode` | `conservative` (default) or `permissive` |
| `--format` | Report format: `text` (default) or `json` |

**Examples**

```bash
# Coerce and write to file
jsonschemaprofiles coerce schema --profile OPENAI_202602 --in schema.json --out fixed.json

# Dry run to see what would change
jsonschemaprofiles coerce schema --profile OPENAI_202602 --in schema.json --dry-run

# Permissive mode (drops unsupported keywords)
jsonschemaprofiles coerce schema --profile OPENAI_202602 --in schema.json --mode permissive
```

## Exit Codes

| Code | Meaning |
|---|---|
| `0` | Schema is compliant |
| `1` | Schema is non-compliant (validation errors) |
| `2` | Internal error (bad profile, parse failure, etc.) |
