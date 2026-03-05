---
layout: default
title: Schemas
nav_order: 4
permalink: /schemas
---

# Schemas
{: .no_toc }

## Table of Contents
{: .no_toc .text-delta }

- TOC
{:toc}

---

Schema profile files live in [`/schemas`](https://github.com/UnitVectorY-Labs/jsonschemaprofiles/tree/main/schemas).  
These profiles validate **JSON Schema documents** against provider requirements.

## Versioning Strategy

Provider docs do not currently publish stable schema-profile versions.  
This project uses a **baseline version** in `YYYYMM` format:

- `202602` means the profile is aligned to the February 2026 requirement baseline.
- When provider requirements change, add new files with a newer baseline (for example `202607`) instead of mutating old files.

## Requirement Reference Enums

Use these enum-like identifiers in code when selecting requirement sets:

- `OPENAI_202602`
- `GEMINI_202602`
- `GEMINI_2_0_202602`
- `MINIMAL_202602`

## Available Profiles And Restrictions

- Gemini restrictions are documented as model/API behavior and are not split by SDK language.
- [openai_202602.yaml](https://github.com/UnitVectorY-Labs/jsonschemaprofiles/blob/main/schemas/openai_202602.yaml): OpenAI Structured Outputs subset. Enforces root object style and keyword subset. Additional runtime checks are still required for depth/property/string/enum limits.
- [gemini_202602.yaml](https://github.com/UnitVectorY-Labs/jsonschemaprofiles/blob/main/schemas/gemini_202602.yaml): Gemini JSON Schema subset from the documented support table.
- [gemini_2_0_202602.yaml](https://github.com/UnitVectorY-Labs/jsonschemaprofiles/blob/main/schemas/gemini_2_0_202602.yaml): Gemini 2.0-focused profile. Same subset as `gemini_202602` plus required `propertyOrdering` for object schemas.
- [minimal_202602.yaml](https://github.com/UnitVectorY-Labs/jsonschemaprofiles/blob/main/schemas/minimal_202602.yaml): Lowest common denominator across OpenAI and Gemini. This is the most portable profile.

## Provider Requirement Pages

- [OpenAI profile details](OPENAI.md)
- [Gemini profile details](GEMINI.md)

## Validation Model

Use a two-phase validation flow:

- Phase 1: validate the candidate schema against one profile meta-schema in [`/schemas`](https://github.com/UnitVectorY-Labs/jsonschemaprofiles/tree/main/schemas).
- Phase 2: run provider-specific code checks for constraints not fully representable in JSON Schema (for example OpenAI global budgets and model-target checks).

Passing only phase 1 is not sufficient for full provider compliance.
