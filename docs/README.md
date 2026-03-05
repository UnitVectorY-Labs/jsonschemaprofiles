---
layout: default
title: jsonschemaprofiles
nav_order: 1
permalink: /
---

# jsonschemaprofiles

A Go library and CLI tool for validating that **JSON Schema documents** conform to provider-specific structured-output restrictions.

This project validates schemas themselves—not JSON data instances—against profiles that capture the documented requirements of providers like OpenAI and Google Gemini.

## Why?

LLM providers that support structured JSON output (OpenAI Structured Outputs, Gemini structured output) each accept only a **subset** of JSON Schema. A schema that works with one provider may be rejected by another. This library:

- Validates schemas against provider profiles using a **two-phase** model (meta-schema + code checks)
- Coerces schemas toward compliance with minimal, traceable changes
- Provides a single library usable by multiple applications
- Offers a CLI for quick validation and testing

## Two-Phase Validation

1. **Phase 1 — Meta-schema validation:** Validates the candidate schema against the profile's embedded meta-schema file using JSON Schema Draft 2020-12.
2. **Phase 2 — Provider rules validation:** Runs provider-specific checks not expressible in JSON Schema (limits, graph traversal, semantic invariants).

## Available Profiles

| Profile ID | Description |
|---|---|
| `OPENAI_202602` | OpenAI Structured Outputs subset |
| `GEMINI_202602` | Gemini baseline structured output |
| `GEMINI_202503` | Gemini 2.0 with required `propertyOrdering` |
| `MINIMAL_202602` | Lowest common denominator across providers |

## Quick Start

### As a Go Library

```go
import jsp "github.com/UnitVectorY-Labs/jsonschemaprofiles"

report, err := jsp.ValidateSchema(jsp.OPENAI_202602, schemaBytes, nil)
if err != nil {
    log.Fatal(err)
}
if !report.Valid {
    fmt.Println(report.Text())
}
```

### As a CLI

```bash
jsonschemaprofiles validate schema --profile OPENAI_202602 --in schema.json
```
