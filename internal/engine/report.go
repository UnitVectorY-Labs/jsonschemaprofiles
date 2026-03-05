package engine

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Severity represents the severity of a finding.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// Finding represents a single validation or coercion finding.
type Finding struct {
	Severity Severity `json:"severity"`
	Code     string   `json:"code"`
	Path     string   `json:"path"`
	Message  string   `json:"message"`
	Hint     string   `json:"hint,omitempty"`
	Rule     string   `json:"rule,omitempty"`
}

// Report is the canonical output of validation or coercion.
type Report struct {
	Valid    bool      `json:"valid"`
	Findings []Finding `json:"findings"`
}

// NewReport creates a new empty report (valid by default).
func NewReport() *Report {
	return &Report{Valid: true, Findings: []Finding{}}
}

// AddFinding adds a finding and updates validity.
func (r *Report) AddFinding(f Finding) {
	r.Findings = append(r.Findings, f)
	if f.Severity == SeverityError {
		r.Valid = false
	}
}

// HasErrors returns true if any finding has error severity.
func (r *Report) HasErrors() bool {
	for _, f := range r.Findings {
		if f.Severity == SeverityError {
			return true
		}
	}
	return false
}

// HasWarnings returns true if any finding has warning severity.
func (r *Report) HasWarnings() bool {
	for _, f := range r.Findings {
		if f.Severity == SeverityWarning {
			return true
		}
	}
	return false
}

// Sort sorts findings by path then code for deterministic output.
func (r *Report) Sort() {
	sort.Slice(r.Findings, func(i, j int) bool {
		if r.Findings[i].Path != r.Findings[j].Path {
			return r.Findings[i].Path < r.Findings[j].Path
		}
		return r.Findings[i].Code < r.Findings[j].Code
	})
}

// JSON returns the JSON encoding of the report.
func (r *Report) JSON() ([]byte, error) {
	r.Sort()
	return json.MarshalIndent(r, "", "  ")
}

// Text returns a human-friendly text representation.
func (r *Report) Text() string {
	r.Sort()
	var b strings.Builder
	if r.Valid {
		b.WriteString("Schema is VALID\n")
	} else {
		b.WriteString("Schema is NOT VALID\n")
	}
	for _, f := range r.Findings {
		b.WriteString(fmt.Sprintf("  [%s] %s: %s", strings.ToUpper(string(f.Severity)), f.Code, f.Message))
		if f.Path != "" {
			b.WriteString(fmt.Sprintf(" (at %s)", f.Path))
		}
		b.WriteString("\n")
		if f.Hint != "" {
			b.WriteString(fmt.Sprintf("    hint: %s\n", f.Hint))
		}
	}
	if len(r.Findings) == 0 {
		b.WriteString("  No findings.\n")
	}
	return b.String()
}
