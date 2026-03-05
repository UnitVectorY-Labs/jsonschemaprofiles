package jsonschemaprofiles

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestReportValid(t *testing.T) {
	r := NewReport()
	if !r.Valid {
		t.Error("new report should be valid")
	}
	if r.HasErrors() {
		t.Error("new report should have no errors")
	}
}

func TestReportAddError(t *testing.T) {
	r := NewReport()
	r.AddFinding(Finding{
		Severity: SeverityError,
		Code:     "TEST_ERROR",
		Path:     "/test",
		Message:  "test error",
	})
	if r.Valid {
		t.Error("report with error should be invalid")
	}
	if !r.HasErrors() {
		t.Error("report should have errors")
	}
}

func TestReportAddWarning(t *testing.T) {
	r := NewReport()
	r.AddFinding(Finding{
		Severity: SeverityWarning,
		Code:     "TEST_WARNING",
		Path:     "/test",
		Message:  "test warning",
	})
	if !r.Valid {
		t.Error("report with only warnings should still be valid")
	}
	if !r.HasWarnings() {
		t.Error("report should have warnings")
	}
}

func TestReportJSON(t *testing.T) {
	r := NewReport()
	r.AddFinding(Finding{
		Severity: SeverityError,
		Code:     "TEST",
		Path:     "/a",
		Message:  "msg",
	})
	data, err := r.JSON()
	if err != nil {
		t.Fatalf("JSON error: %v", err)
	}

	var parsed Report
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if parsed.Valid {
		t.Error("parsed report should be invalid")
	}
	if len(parsed.Findings) != 1 {
		t.Errorf("expected 1 finding, got %d", len(parsed.Findings))
	}
}

func TestReportText(t *testing.T) {
	r := NewReport()
	r.AddFinding(Finding{
		Severity: SeverityError,
		Code:     "TEST",
		Path:     "/a",
		Message:  "msg",
		Hint:     "fix it",
	})
	text := r.Text()
	if !strings.Contains(text, "NOT VALID") {
		t.Error("text should contain NOT VALID")
	}
	if !strings.Contains(text, "TEST") {
		t.Error("text should contain error code")
	}
	if !strings.Contains(text, "fix it") {
		t.Error("text should contain hint")
	}
}

func TestReportSort(t *testing.T) {
	r := NewReport()
	r.AddFinding(Finding{Severity: SeverityError, Code: "B", Path: "/z", Message: "b"})
	r.AddFinding(Finding{Severity: SeverityError, Code: "A", Path: "/a", Message: "a"})
	r.AddFinding(Finding{Severity: SeverityError, Code: "A", Path: "/z", Message: "a2"})
	r.Sort()

	if r.Findings[0].Path != "/a" {
		t.Error("first finding should be at /a")
	}
	if r.Findings[1].Code != "A" || r.Findings[1].Path != "/z" {
		t.Error("second finding should be A at /z")
	}
	if r.Findings[2].Code != "B" {
		t.Error("third finding should be B")
	}
}
