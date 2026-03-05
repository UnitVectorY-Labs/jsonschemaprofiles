package jsonschemaprofiles

import (
	"encoding/json"
	"testing"
)

func TestCoerceOpenAIAddsMissingFields(t *testing.T) {
	// Schema missing additionalProperties and required
	input := `{
		"type": "object",
		"properties": {
			"name": { "type": "string" }
		}
	}`

	coerced, report, changed, err := CoerceSchema(OPENAI_202602, []byte(input), &CoerceOptions{Mode: CoerceModeConservative})
	if err != nil {
		t.Fatalf("CoerceSchema error: %v", err)
	}
	if !changed {
		t.Error("expected changes but got none")
	}

	var result map[string]interface{}
	if err := json.Unmarshal(coerced, &result); err != nil {
		t.Fatalf("failed to parse coerced schema: %v", err)
	}

	// Should have additionalProperties: false
	if ap, ok := result["additionalProperties"]; !ok || ap != false {
		t.Error("expected additionalProperties: false")
	}

	// Should have required: ["name"]
	req, ok := result["required"]
	if !ok {
		t.Error("expected required field")
	}
	if reqArr, ok := req.([]interface{}); ok {
		if len(reqArr) != 1 || reqArr[0] != "name" {
			t.Errorf("expected required: [\"name\"], got: %v", reqArr)
		}
	}

	// Report should have findings
	if len(report.Findings) == 0 {
		t.Error("expected coercion findings")
	}
}

func TestCoerceGemini20AddsPropertyOrdering(t *testing.T) {
	input := `{
		"type": "object",
		"properties": {
			"name": { "type": "string" },
			"age": { "type": "integer" }
		},
		"required": ["name", "age"]
	}`

	coerced, report, changed, err := CoerceSchema(GEMINI_202503, []byte(input), &CoerceOptions{Mode: CoerceModeConservative})
	if err != nil {
		t.Fatalf("CoerceSchema error: %v", err)
	}
	if !changed {
		t.Error("expected changes but got none")
	}

	var result map[string]interface{}
	if err := json.Unmarshal(coerced, &result); err != nil {
		t.Fatalf("failed to parse coerced schema: %v", err)
	}

	// Should have propertyOrdering
	if _, ok := result["propertyOrdering"]; !ok {
		t.Error("expected propertyOrdering to be added")
	}

	// Report should document the change
	found := false
	for _, f := range report.Findings {
		if f.Code == "COERCE_ADD_PROPERTY_ORDERING" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected COERCE_ADD_PROPERTY_ORDERING finding")
	}
}

func TestCoerceDryRunDoesNotModify(t *testing.T) {
	input := `{
		"type": "object",
		"properties": {
			"name": { "type": "string" }
		}
	}`

	coerced, report, changed, err := CoerceSchema(OPENAI_202602, []byte(input), &CoerceOptions{Mode: CoerceModeConservative, DryRun: true})
	if err != nil {
		t.Fatalf("CoerceSchema error: %v", err)
	}
	if !changed {
		t.Error("expected changes flag to be true even in dry-run")
	}

	// Coerced bytes should be the same as input
	if string(coerced) != input {
		t.Error("dry-run should return original bytes unchanged")
	}

	// Report should still have findings
	if len(report.Findings) == 0 {
		t.Error("expected coercion findings in dry-run mode")
	}
}

func TestCoerceModeOff(t *testing.T) {
	input := `{
		"type": "object",
		"properties": {
			"name": { "type": "string" }
		}
	}`

	_, report, changed, err := CoerceSchema(OPENAI_202602, []byte(input), &CoerceOptions{Mode: CoerceModeOff})
	if err != nil {
		t.Fatalf("CoerceSchema error: %v", err)
	}
	if changed {
		t.Error("expected no changes in off mode")
	}

	// Report should indicate validation errors
	if report.Valid {
		t.Error("expected invalid report in off mode since schema is not compliant")
	}
}

func TestCoercePermissiveDropsKeywords(t *testing.T) {
	input := `{
		"type": "object",
		"properties": {
			"name": { "type": "string" }
		},
		"required": ["name"],
		"additionalProperties": false,
		"allOf": [{ "type": "object" }]
	}`

	coerced, report, changed, err := CoerceSchema(OPENAI_202602, []byte(input), &CoerceOptions{Mode: CoerceModePermissive})
	if err != nil {
		t.Fatalf("CoerceSchema error: %v", err)
	}
	if !changed {
		t.Error("expected changes")
	}

	var result map[string]interface{}
	if err := json.Unmarshal(coerced, &result); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if _, has := result["allOf"]; has {
		t.Error("expected allOf to be removed in permissive mode")
	}

	// Check for warning finding
	found := false
	for _, f := range report.Findings {
		if f.Code == "COERCE_DROP_UNSUPPORTED_KEYWORD" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected COERCE_DROP_UNSUPPORTED_KEYWORD finding")
	}
	if !report.Valid {
		t.Error("expected report to be valid after permissive coercion")
	}
}

func TestCoerceConservativeReportsInvalidWhenSchemaRemainsNonCompliant(t *testing.T) {
	input := `{"type":"object","properties":{"value":{"type":"string","if":{"const":"x"},"then":{"type":"string"}}},"required":["value"],"additionalProperties":false}`

	coerced, report, changed, err := CoerceSchema(OPENAI_202602, []byte(input), &CoerceOptions{Mode: CoerceModeConservative})
	if err != nil {
		t.Fatalf("CoerceSchema error: %v", err)
	}
	if changed {
		t.Error("expected no coercion changes for unsupported keywords in conservative mode")
	}
	if string(coerced) != input {
		t.Error("expected unchanged bytes when no changes are applied")
	}
	if report.Valid {
		t.Error("expected invalid report when schema remains non-compliant after coercion")
	}

	foundMetaValidation := false
	for _, f := range report.Findings {
		if f.Code == "META_SCHEMA_VALIDATION" {
			foundMetaValidation = true
			break
		}
	}
	if !foundMetaValidation {
		t.Error("expected meta-schema validation finding in coercion report")
	}
}

func TestCoerceNoChangesReturnsOriginalBytes(t *testing.T) {
	input := `{"type":"object","properties":{"name":{"type":"string"}},"required":["name"],"additionalProperties":false}`

	coerced, report, changed, err := CoerceSchema(OPENAI_202602, []byte(input), &CoerceOptions{Mode: CoerceModeConservative})
	if err != nil {
		t.Fatalf("CoerceSchema error: %v", err)
	}
	if changed {
		t.Error("expected no changes for already compliant schema")
	}
	if !report.Valid {
		t.Error("expected valid report for already compliant schema")
	}
	if string(coerced) != input {
		t.Error("expected original bytes to be returned when no changes are applied")
	}
}
