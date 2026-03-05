package jsonschemaprofiles

import (
	"embed"
	"path/filepath"
	"strings"
	"testing"
)

//go:embed tests
var testFS embed.FS

// profileDirMap maps directory names to ProfileIDs
var profileDirMap = map[string]ProfileID{
	"openai_202602":  OPENAI_202602,
	"gemini_202602":  GEMINI_202602,
	"gemini_202503":  GEMINI_202503,
	"minimal_202602": MINIMAL_202602,
}

func TestValidSchemas(t *testing.T) {
	for dirName, profileID := range profileDirMap {
		validDir := filepath.Join("tests", dirName, "valid")
		entries, err := testFS.ReadDir(validDir)
		if err != nil {
			t.Fatalf("failed to read valid test dir %s: %v", validDir, err)
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}
			name := entry.Name()
			t.Run(string(profileID)+"/valid/"+name, func(t *testing.T) {
				data, err := testFS.ReadFile(filepath.Join(validDir, name))
				if err != nil {
					t.Fatalf("failed to read test file: %v", err)
				}
				report, err := ValidateSchema(profileID, data, nil)
				if err != nil {
					t.Fatalf("ValidateSchema returned error: %v", err)
				}
				if !report.Valid {
					t.Errorf("expected valid schema but got invalid.\nFindings:")
					for _, f := range report.Findings {
						t.Errorf("  [%s] %s: %s (at %s)", f.Severity, f.Code, f.Message, f.Path)
					}
				}
			})
		}
	}
}

func TestInvalidSchemas(t *testing.T) {
	for dirName, profileID := range profileDirMap {
		invalidDir := filepath.Join("tests", dirName, "invalid")
		entries, err := testFS.ReadDir(invalidDir)
		if err != nil {
			t.Fatalf("failed to read invalid test dir %s: %v", invalidDir, err)
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}
			name := entry.Name()
			t.Run(string(profileID)+"/invalid/"+name, func(t *testing.T) {
				data, err := testFS.ReadFile(filepath.Join(invalidDir, name))
				if err != nil {
					t.Fatalf("failed to read test file: %v", err)
				}
				report, err := ValidateSchema(profileID, data, nil)
				if err != nil {
					t.Fatalf("ValidateSchema returned error: %v", err)
				}
				if report.Valid {
					t.Errorf("expected invalid schema but got valid (no errors found)")
				}
			})
		}
	}
}
