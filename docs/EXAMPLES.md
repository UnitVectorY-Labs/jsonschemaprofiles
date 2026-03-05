---
layout: default
title: Examples
nav_order: 5
permalink: /examples
---

# Examples
{: .no_toc }

## Table of Contents
{: .no_toc .text-delta }

- TOC
{:toc}

---

## OpenAI: Schema Fails → Coercion Fixes → Schema Passes

### Starting Schema

A schema missing `additionalProperties` and `required`:

```json
{
  "type": "object",
  "properties": {
    "name": { "type": "string" },
    "age": { "type": "integer" }
  }
}
```

### Step 1: Validate

```bash
$ jsonschemaprofiles validate schema --profile OPENAI_202602 --in schema.json
Schema is NOT VALID
  [ERROR] META_SCHEMA_VALIDATION: missing properties 'required', 'additionalProperties'
  [ERROR] OPENAI_MISSING_ADDITIONAL_PROPERTIES: Object schema must have additionalProperties: false
  [ERROR] OPENAI_MISSING_REQUIRED: Object with properties must have required listing all property keys
```

### Step 2: Coerce

```bash
$ jsonschemaprofiles coerce schema --profile OPENAI_202602 --in schema.json --out fixed.json
Schema is VALID
  [INFO] COERCE_ADD_ADDITIONAL_PROPERTIES_FALSE: Added additionalProperties: false
  [INFO] COERCE_ADD_REQUIRED: Added required with all property keys
```

### Step 3: Validate the Fixed Schema

```bash
$ jsonschemaprofiles validate schema --profile OPENAI_202602 --in fixed.json
Schema is VALID
  No findings.
```

The coerced `fixed.json`:

```json
{
  "additionalProperties": false,
  "properties": {
    "age": { "type": "integer" },
    "name": { "type": "string" }
  },
  "required": ["age", "name"],
  "type": "object"
}
```

---

## Gemini 2.0: Adding propertyOrdering

### Starting Schema

```json
{
  "type": "object",
  "properties": {
    "title": { "type": "string" },
    "rating": { "type": "number", "minimum": 0, "maximum": 5 }
  },
  "required": ["title", "rating"]
}
```

### Validate Against Gemini 2.0

```bash
$ jsonschemaprofiles validate schema --profile GEMINI_2_0_202602 --in schema.json
Schema is NOT VALID
  [ERROR] GEMINI_2_0_MISSING_PROPERTY_ORDERING: Object schema must include propertyOrdering for Gemini 2.0 compatibility
```

### Coerce for Gemini 2.0

```bash
$ jsonschemaprofiles coerce schema --profile GEMINI_2_0_202602 --in schema.json
{
  "properties": {
    "rating": { "maximum": 5, "minimum": 0, "type": "number" },
    "title": { "type": "string" }
  },
  "propertyOrdering": ["rating", "title"],
  "required": ["title", "rating"],
  "type": "object"
}
Schema is VALID
  [WARNING] COERCE_ADD_PROPERTY_ORDERING: Added propertyOrdering derived from properties keys (ordering may not match source)
```

---

## Minimal Profile: Cross-Provider Compatibility

### Starting Schema

```json
{
  "type": "object",
  "properties": {
    "message": { "type": "string" }
  }
}
```

### Validate Against Minimal

```bash
$ jsonschemaprofiles validate schema --profile MINIMAL_202602 --in schema.json
Schema is NOT VALID
  [ERROR] MINIMAL_MISSING_ADDITIONAL_PROPERTIES: Object schema must have additionalProperties: false
  [ERROR] MINIMAL_MISSING_REQUIRED: Object with properties must have required listing all property keys
```

### Coerce for Maximum Compatibility

```bash
$ jsonschemaprofiles coerce schema --profile MINIMAL_202602 --in schema.json --out portable.json
```

---

## Go Library Usage

### Validate a Schema

```go
package main

import (
    "fmt"
    "log"

    jsp "github.com/UnitVectorY-Labs/jsonschemaprofiles"
)

func main() {
    schema := []byte(`{
        "type": "object",
        "properties": {
            "name": { "type": "string" }
        },
        "required": ["name"],
        "additionalProperties": false
    }`)

    report, err := jsp.ValidateSchema(jsp.OPENAI_202602, schema, nil)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Print(report.Text())
    // Output: Schema is VALID
    //   No findings.
}
```

### Coerce a Schema

```go
coerced, report, changed, err := jsp.CoerceSchema(jsp.OPENAI_202602, inputSchema, &jsp.CoerceOptions{
    Mode: jsp.CoerceModeConservative,
})
if err != nil {
    log.Fatal(err)
}
if changed {
    fmt.Println("Schema was modified:")
    fmt.Println(string(coerced))
}
for _, f := range report.Findings {
    fmt.Printf("[%s] %s: %s\n", f.Severity, f.Code, f.Message)
}
```

### List Profiles Programmatically

```go
for _, p := range jsp.ListProfiles() {
    fmt.Printf("%-25s %s (baseline %s)\n", p.ID, p.Title, p.Baseline)
}
```
