# jsonschemaprofiles

A Go library and CLI tool for validating that **JSON Schema documents** conform to provider-specific structured-output restrictions (for example Gemini and OpenAI).

This project is about **schema-profile validation**—it validates schemas themselves, not JSON data instances.

## Overview

LLM providers that support structured JSON output each accept only a subset of JSON Schema. A schema that works with one provider may be rejected by another. This library:

- Validates schemas against provider profiles using a two-phase model
- Coerces schemas toward compliance with minimal, traceable changes
- Embeds all profile meta-schemas for zero-config usage
- Provides both a Go library API and a CLI

## Installation

### Go Library

```bash
go get github.com/UnitVectorY-Labs/jsonschemaprofiles
```

### CLI

```bash
go install github.com/UnitVectorY-Labs/jsonschemaprofiles/cmd/jsonschemaprofiles@latest
```

Or download a pre-built binary from [Releases](https://github.com/UnitVectorY-Labs/jsonschemaprofiles/releases).

## Quick Start

### Library

```go
import jsp "github.com/UnitVectorY-Labs/jsonschemaprofiles"

// Validate a schema against a profile
report, err := jsp.ValidateSchema(jsp.OPENAI_202602, schemaBytes, nil)
if !report.Valid {
    fmt.Println(report.Text())
}

// Coerce a schema for compliance
coerced, report, changed, err := jsp.CoerceSchema(jsp.OPENAI_202602, schemaBytes, &jsp.CoerceOptions{
    Mode: jsp.CoerceModeConservative,
})
```

### CLI

```bash
# List profiles
jsonschemaprofiles profiles list

# Validate
jsonschemaprofiles validate schema --profile OPENAI_202602 --in schema.json

# Coerce
jsonschemaprofiles coerce schema --profile OPENAI_202602 --in schema.json --out fixed.json
```

## Available Profiles

| Profile ID | Description |
|---|---|
| `OPENAI_202602` | OpenAI Structured Outputs subset |
| `GEMINI_202602` | Gemini baseline structured output |
| `GEMINI_2_0_202602` | Gemini 2.0 with required `propertyOrdering` |
| `MINIMAL_202602` | Lowest common denominator across providers |

## Documentation

Full documentation is available at [jsonschemaprofiles.dev](https://jsonschemaprofiles.dev):

- [Installation](https://jsonschemaprofiles.dev/install)
- [Usage](https://jsonschemaprofiles.dev/usage)
- [Library API](https://jsonschemaprofiles.dev/library)
- [Examples](https://jsonschemaprofiles.dev/examples)
- [Schemas & Profiles](https://jsonschemaprofiles.dev/schemas)
- [OpenAI Profile Details](https://jsonschemaprofiles.dev/openai)
- [Gemini Profile Details](https://jsonschemaprofiles.dev/gemini)

## License

See [LICENSE](LICENSE) for details.
