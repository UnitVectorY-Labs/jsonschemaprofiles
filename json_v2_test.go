package jsonschemaprofiles

import (
	"strings"
	"testing"
)

func TestSchemaJSONRejectsAmbiguousInput(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{"duplicate root name", []byte(`{"type":"object","type":"string"}`)},
		{"duplicate nested name", []byte(`{"properties":{"x":{"type":"string","type":"number"}}}`)},
		{"invalid UTF-8", []byte("{\"description\":\"\xff\"}")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ValidateSchema(OPENAI_202602, tt.input, nil); err == nil {
				t.Error("validation accepted invalid JSON")
			}
			for _, mode := range []CoerceMode{CoerceModeConservative, CoerceModeOff} {
				if _, _, _, err := CoerceSchema(OPENAI_202602, tt.input, &CoerceOptions{Mode: mode}); err == nil {
					t.Errorf("coercion mode %s accepted invalid JSON", mode)
				}
			}
		})
	}
}

func TestReportJSONEmptyFindings(t *testing.T) {
	r := &Report{Valid: true}
	data, err := r.JSON()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"findings": []`) {
		t.Errorf("expected empty findings array: %s", data)
	}
}
