---
layout: default
title: Library API
nav_order: 5
permalink: /library
---

# Library API
{: .no_toc }

## Table of Contents
{: .no_toc .text-delta }

- TOC
{:toc}

---

This page contains the Go library reference for `github.com/UnitVectorY-Labs/jsonschemaprofiles`, including usage examples and the complete public API.

Implementation details under `internal/` are intentionally excluded from this interface and are not importable by external modules.

## Installation

```bash
go get github.com/UnitVectorY-Labs/jsonschemaprofiles
```

## Import

```go
import jsp "github.com/UnitVectorY-Labs/jsonschemaprofiles"
```

## Common Usage

### Profile Discovery

```go
// List all available profiles
profiles := jsp.ListProfiles()
for _, p := range profiles {
    fmt.Printf("%s: %s\n", p.ID, p.Title)
}

// Get info for one profile
info, err := jsp.GetProfileInfo(jsp.OPENAI_202602)

// Get embedded YAML meta-schema bytes
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

### Report Output

```go
jsonBytes, _ := report.JSON() // Sorted, indented JSON
textOutput := report.Text()   // Human-friendly text
```

## Public API Reference

### Profiles

#### `type ProfileID string`

Profile identifiers:

- `OPENAI_202602`
- `GEMINI_202602`
- `GEMINI_2_0_202602`
- `MINIMAL_202602`

#### `type ProfileInfo struct`

```go
type ProfileInfo struct {
    ID          ProfileID
    Title       string
    Description string
    Baseline    string
    SchemaFile  string
}
```

`SchemaFile` is the path (within embedded assets) to the profile meta-schema YAML.

#### `func ListProfiles() []ProfileInfo`

Returns all registered profiles in stable order.

#### `func GetProfileInfo(id ProfileID) (ProfileInfo, error)`

Returns profile metadata for one profile ID.

#### `func GetProfileSchema(id ProfileID) ([]byte, error)`

Returns embedded YAML meta-schema bytes for a profile.

### Validation

#### `type ValidateOptions struct`

```go
type ValidateOptions struct {
    Strict      bool
    ModelTarget string
}
```

- `Strict`: promotes warnings to errors.
- `ModelTarget`: optional target-specific behavior (for example `"fine-tuned"`).

#### `func ValidateSchema(profileID ProfileID, schemaBytes []byte, opts *ValidateOptions) (*Report, error)`

Validates a candidate schema against a profile and returns a `Report`.

### Coercion

#### `type CoerceMode string`

Supported modes:

- `CoerceModeConservative`
- `CoerceModePermissive`
- `CoerceModeOff`

#### `type CoerceOptions struct`

```go
type CoerceOptions struct {
    Mode   CoerceMode
    DryRun bool
}
```

#### `func CoerceSchema(profileID ProfileID, schemaBytes []byte, opts *CoerceOptions) ([]byte, *Report, bool, error)`

Attempts profile-specific schema coercion.

Return values:

1. Coerced schema bytes.
2. Coercion report.
3. `changed` flag.
4. Error.

### Reporting

#### `type Severity string`

Supported values:

- `SeverityError`
- `SeverityWarning`
- `SeverityInfo`

#### `type Finding struct`

```go
type Finding struct {
    Severity Severity `json:"severity"`
    Code     string   `json:"code"`
    Path     string   `json:"path"`
    Message  string   `json:"message"`
    Hint     string   `json:"hint,omitempty"`
    Rule     string   `json:"rule,omitempty"`
}
```

#### `type Report struct`

```go
type Report struct {
    Valid    bool      `json:"valid"`
    Findings []Finding `json:"findings"`
}
```

#### `func NewReport() *Report`

Creates an empty report (`Valid=true`).

#### Report Methods

- `func (r *Report) AddFinding(f Finding)`
- `func (r *Report) HasErrors() bool`
- `func (r *Report) HasWarnings() bool`
- `func (r *Report) Sort()`
- `func (r *Report) JSON() ([]byte, error)`
- `func (r *Report) Text() string`
