---
layout: default
parent: Schemas
title: OpenAI
nav_order: 5
permalink: /schemas/openai
---

# OpenAI Structured Outputs Profile
{: .no_toc }

## Table of Contents
{: .no_toc .text-delta }

- TOC
{:toc}

---

This page captures the JSON Schema subset and constraints documented by OpenAI for Structured Outputs.

## Source

- OpenAI docs: [Structured Outputs - Supported schemas](https://developers.openai.com/api/docs/guides/structured-outputs/#supported-schemas)
- Profile baseline version: **202602** (February 2026)
- Notes on dating: OpenAI's page does not expose a clear per-section "last updated" timestamp in the same way as Gemini, so this repository normalizes to the February 2026 baseline window for cross-provider profile alignment.

## Documented Requirements

### Supported types

- `string`
- `number`
- `boolean`
- `integer`
- `object`
- `array`
- `enum`
- `anyOf` (nested, not at root)

### Supported constraints (documented)

- `string`: `pattern`, `format`
- `number`: `multipleOf`, `maximum`, `exclusiveMaximum`, `minimum`, `exclusiveMinimum`
- `array`: `minItems`, `maxItems`
- Objects are represented with `properties`, `required`, and `additionalProperties`

### Hard constraints

- Root schema must be an `object`
- Root schema must not be `anyOf`
- All fields must be listed in `required`
- `additionalProperties: false` must be set for objects
- Optional behavior is represented as union-with-null, such as `type: ["string", "null"]`

### Limits

- Up to 5,000 object properties total
- Up to 10 nested object levels
- Total string length across property names, definition names, enum values, and const values <= 120,000 characters
- Up to 1,000 enum values across the schema
- If one string enum has more than 250 values, combined enum string length must be <= 15,000 characters

### Unsupported keywords (documented)

- Composition keywords: `allOf`, `not`, `dependentRequired`, `dependentSchemas`, `if`, `then`, `else`

### Fine-tuned models (additional documented unsupported keywords)

- `string`: `minLength`, `maxLength`, `pattern`, `format`
- `number`: `minimum`, `maximum`, `multipleOf`
- `object`: `patternProperties`, `unevaluatedProperties`, `propertyNames`, `minProperties`, `maxProperties`
- `array`: `unevaluatedItems`, `contains`, `minContains`, `maxContains`, `minItems`, `maxItems`, `uniqueItems`

### Documented supported advanced features

- `$defs` definitions
- `$ref` references
- Recursive schemas

## Profile File

- [openai_202602.yaml](../schemas/openai_202602.yaml)

## Requirement Reference Enum

- `OPENAI_202602`

## Notes

- The profile schema enforces the documented subset that is practical to encode in JSON Schema itself.
- Some published limits (for example total string lengths and total nested-property counts) are runtime constraints and are documented here but are not fully representable as static JSON Schema assertions.

## Validation That Must Run In Code

Even when a schema passes [openai_202602.yaml](../schemas/openai_202602.yaml), additional checks are required in application code:

- Count total object properties across the full schema graph and enforce the `<= 5000` limit.
- Measure object nesting depth and enforce the `<= 10` level limit.
- Compute cumulative string budget across property names, `$defs` names, enum values, and const values, then enforce `<= 120000` characters.
- Count total enum values across the schema and enforce `<= 1000`.
- For any single string enum with more than 250 values, enforce combined enum string length `<= 15000`.
- Enforce exact equality between each object's `properties` keys and `required` keys (no missing or extra keys in either set).
- If targeting OpenAI fine-tuned models, apply the additional model-specific unsupported-keyword restrictions documented above.

The profile meta-schema should be treated as phase one. Provider limit checks in code are phase two.
