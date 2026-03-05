---
layout: default
parent: Schemas
title: Gemini
nav_order: 6
permalink: /schemas/gemini
---

# Gemini Structured Output Profile
{: .no_toc }

## Table of Contents
{: .no_toc .text-delta }

- TOC
{:toc}

---

This page captures the JSON Schema subset documented by Google for Gemini structured output.

## Source

- Google Gemini docs: [Structured output - JSON Schema support](https://ai.google.dev/gemini-api/docs/structured-output?example=recipe#json_schema_support)
- Profile baseline version: **202602** (February 2026)
- Upstream page update shown by Google: **February 26, 2026 (UTC)**

## Documented Requirements

### Supported schema keyword groups

- Type values: `string`, `integer`, `number`, `boolean`, `array`, `object`, `null`
- Generic fields: `title`, `description`
- `string`: `enum`, `format`
- `integer` and `number`: `format`, `minimum`, `maximum`, `enum`
- `array`: `minItems`, `maxItems`, `items`, `prefixItems`
- `object`: `properties`, `required`, `propertyOrdering`, `additionalProperties`

### Nullability

- A nullable field uses a type array containing `null`, for example `type: ["string", "null"]`.

### Model behavior note for property order

- The docs state that Gemini API and SDK preserve schema property ordering.
- For **Gemini 2.0 and earlier**, examples and schemas should include `propertyOrdering[]`.

### Language vs model differences

- The documented JSON Schema subset is API-level and is not different by SDK language.
- SDKs can provide different helpers (for example class/type wrappers), but the transmitted schema restrictions are the same.
- The documented difference is model-generation behavior (not SDK language), specifically the Gemini 2.0 `propertyOrdering` requirement note above.

## Profile Files

- [gemini_202602.yaml](../schemas/gemini_202602.yaml): baseline Gemini documented subset.
- [gemini_2_0_202602.yaml](../schemas/gemini_2_0_202602.yaml): stricter Gemini 2.0 profile that requires `propertyOrdering` on object schemas.

## Requirement Reference Enums

- `GEMINI_202602`
- `GEMINI_2_0_202602`

## Notes

- The profile schema encodes only the subset explicitly documented on the linked page.
- This profile intentionally stays conservative where the page does not explicitly document additional JSON Schema keywords.

## Validation That Must Run In Code

Even when a schema passes [gemini_202602.yaml](../schemas/gemini_202602.yaml), additional checks are still useful in application code:

- If targeting Gemini 2.0 or earlier, enforce project policy that object schemas include `propertyOrdering`.
- When `propertyOrdering` is present, verify every name exists in `properties` and that no duplicate names are present.
- Validate semantic consistency between `required` and `properties` (every required field is declared as a property).
- Apply model-target checks at runtime because support details can differ by model generation even within the documented keyword subset.

For Gemini 2.0 targets, prefer validating against [gemini_2_0_202602.yaml](../schemas/gemini_2_0_202602.yaml) first, then run model-target checks in code.
