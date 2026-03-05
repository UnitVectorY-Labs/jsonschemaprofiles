package jsonschemaprofiles

import "github.com/UnitVectorY-Labs/jsonschemaprofiles/internal/engine"

// Severity represents the severity of a finding.
type Severity = engine.Severity

const (
	SeverityError   Severity = engine.SeverityError
	SeverityWarning Severity = engine.SeverityWarning
	SeverityInfo    Severity = engine.SeverityInfo
)

// Finding represents a single validation or coercion finding.
type Finding = engine.Finding

// Report is the canonical output of validation or coercion.
type Report = engine.Report

// NewReport creates a new empty report (valid by default).
func NewReport() *Report {
	return engine.NewReport()
}
