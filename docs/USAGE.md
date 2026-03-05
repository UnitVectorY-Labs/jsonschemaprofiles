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

## Go Library

### Installation

```bash
go get github.com/UnitVectorY-Labs/jsonschemaprofiles
```

### Profile Discovery

```go
import jsp "github.com/UnitVectorY-Labs/jsonschemaprofiles"

// List all available profiles
profiles := jsp.ListProfiles()
for _, p := range profiles {
    fmt.Printf("%s: %s\n", p.ID, p.Title)
}

// Get info for a specific profile
info, err := jsp.GetProfileInfo(jsp.OPENAI_202602)

// Get the raw YAML meta-schema bytes
schemaBytes, err := jsp.GetProfileSchema(jsp.OPENAI_202602)
```

### Schema Validation

```go
report, err := jsp.ValidateSchema(jsp.OPENAI_202602, schemaBytes, nil)
if err != nil {
    // Internal error (bad profile, parse failure, etc.)
    log.Fatal(err)
}
if !report.Valid {
    fmt.Println(report.Text())
}
```

#### Validation Options

| Option | Type | Description |
|---|---|---|
| `Strict` | `bool` | Treat warnings as errors |
| `ModelTarget` | `string` | Apply model-specific restrictions (e.g., `"fine-tuned"` for OpenAI) |

```go
opts := &jsp.ValidateOptions{
    Strict:      true,
    ModelTarget: "fine-tuned",
}
report, err := jsp.ValidateSchema(jsp.OPENAI_202602, schemaBytes, opts)
```

### Schema Coercion

```go
coerced, report, changed, err := jsp.CoerceSchema(jsp.OPENAI_202602, schemaBytes, &jsp.CoerceOptions{
    Mode: jsp.CoerceModeConservative,
})
if err != nil {
    log.Fatal(err)
}
if changed {
    // coerced contains the modified schema bytes
    os.Stdout.Write(coerced)
}
// report contains details of all applied changes
fmt.Println(report.Text())
```

#### Coercion Modes

| Mode | Description |
|---|---|
| `conservative` | Only add missing required structural fields (default) |
| `permissive` | May drop unsupported keywords when safe, emitting warnings |
| `off` | Validation only, no coercion |

#### Dry Run

```go
coerced, report, changed, err := jsp.CoerceSchema(jsp.OPENAI_202602, schemaBytes, &jsp.CoerceOptions{
    Mode:   jsp.CoerceModeConservative,
    DryRun: true,
})
// coerced == original bytes (unchanged)
// report describes what WOULD change
// changed indicates if changes would be needed
```

### Report Format

The `Report` type is used by both validation and coercion:

```go
type Report struct {
    Valid    bool      `json:"valid"`
    Findings []Finding `json:"findings"`
}

type Finding struct {
    Severity Severity `json:"severity"`     // "error", "warning", "info"
    Code     string   `json:"code"`         // Stable error code
    Path     string   `json:"path"`         // JSON Pointer
    Message  string   `json:"message"`      // Human-readable
    Hint     string   `json:"hint,omitempty"`   // Fix suggestion
    Rule     string   `json:"rule,omitempty"`   // Coercion rule ID
}
```

Reports can be serialized:

```go
jsonBytes, _ := report.JSON()   // Sorted, indented JSON
textOutput := report.Text()     // Human-friendly text
```

---

## CLI

### Commands

#### `profiles list`

List all available profile IDs and their descriptions.

```bash
jsonschemaprofiles profiles list
```

#### `validate schema`

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

##### Examples

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

#### `coerce schema`

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

##### Examples

```bash
# Coerce and write to file
jsonschemaprofiles coerce schema --profile OPENAI_202602 --in schema.json --out fixed.json

# Dry run to see what would change
jsonschemaprofiles coerce schema --profile OPENAI_202602 --in schema.json --dry-run

# Permissive mode (drops unsupported keywords)
jsonschemaprofiles coerce schema --profile OPENAI_202602 --in schema.json --mode permissive
```

### Exit Codes

| Code | Meaning |
|---|---|
| `0` | Schema is compliant |
| `1` | Schema is non-compliant (validation errors) |
| `2` | Internal error (bad profile, parse failure, etc.) |
