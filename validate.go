package jsonschemaprofiles

import (
	"encoding/json"
	"fmt"
	"strings"

	jschema "github.com/santhosh-tekuri/jsonschema/v6"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"gopkg.in/yaml.v3"
)

// printer is used by the jsonschema library to format error messages.
var printer = message.NewPrinter(language.English)

// ValidateOptions configures validation behavior.
type ValidateOptions struct {
	// Strict treats warnings as errors.
	Strict bool
	// ModelTarget applies additional model-specific restrictions (e.g. OpenAI fine-tuned).
	ModelTarget string
}

// ValidateSchema validates a candidate JSON Schema against a profile.
// It runs Phase 1 (meta-schema validation) and Phase 2 (provider-specific code checks).
func ValidateSchema(profileID ProfileID, schemaBytes []byte, opts *ValidateOptions) (*Report, error) {
	if opts == nil {
		opts = &ValidateOptions{}
	}

	report := NewReport()

	// Parse candidate schema
	var candidate interface{}
	if err := json.Unmarshal(schemaBytes, &candidate); err != nil {
		return nil, fmt.Errorf("failed to parse candidate schema as JSON: %w", err)
	}

	// Phase 1: meta-schema validation
	if err := validatePhase1(profileID, candidate, report); err != nil {
		return nil, fmt.Errorf("phase 1 validation error: %w", err)
	}

	// Phase 2: provider-specific code checks
	// Parse into ordered map for traversal
	var candidateMap map[string]interface{}
	if err := json.Unmarshal(schemaBytes, &candidateMap); err != nil {
		// Not an object at top level - phase 2 will catch this
		candidateMap = nil
	}

	validatePhase2(profileID, candidateMap, schemaBytes, report, opts)

	// If strict mode, promote warnings to errors
	if opts.Strict {
		for i := range report.Findings {
			if report.Findings[i].Severity == SeverityWarning {
				report.Findings[i].Severity = SeverityError
				report.Valid = false
			}
		}
	}

	return report, nil
}

// validatePhase1 validates the candidate schema against the profile's meta-schema.
func validatePhase1(profileID ProfileID, candidate interface{}, report *Report) error {
	// Load the profile meta-schema YAML
	yamlBytes, err := GetProfileSchema(profileID)
	if err != nil {
		return err
	}

	// Parse YAML to generic structure
	var metaSchemaObj interface{}
	if err := yaml.Unmarshal(yamlBytes, &metaSchemaObj); err != nil {
		return fmt.Errorf("failed to parse profile meta-schema YAML: %w", err)
	}

	// Convert YAML types to JSON-compatible types
	metaSchemaObj = yamlToJSON(metaSchemaObj)

	// Get the profile info for the schema URL
	info, err := GetProfileInfo(profileID)
	if err != nil {
		return err
	}

	// Create a compiler and add the meta-schema
	c := jschema.NewCompiler()

	schemaURL := "file:///" + info.SchemaFile
	if err := c.AddResource(schemaURL, metaSchemaObj); err != nil {
		return fmt.Errorf("failed to add meta-schema resource: %w", err)
	}

	compiled, err := c.Compile(schemaURL)
	if err != nil {
		return fmt.Errorf("failed to compile meta-schema: %w", err)
	}

	// Validate the candidate against the compiled meta-schema
	validationErr := compiled.Validate(candidate)
	if validationErr != nil {
		// Extract validation errors from the library
		if ve, ok := validationErr.(*jschema.ValidationError); ok {
			addValidationErrors(ve, report, "")
		} else {
			report.AddFinding(Finding{
				Severity: SeverityError,
				Code:     "META_SCHEMA_VALIDATION_FAILED",
				Path:     "",
				Message:  validationErr.Error(),
			})
		}
	}

	return nil
}

// addValidationErrors recursively extracts errors from the jsonschema validation error tree.
func addValidationErrors(ve *jschema.ValidationError, report *Report, parentPath string) {
	path := instanceLocationToPath(ve.InstanceLocation)
	if path == "" {
		path = parentPath
	}

	if len(ve.Causes) == 0 {
		msg := ve.ErrorKind.LocalizedString(printer)
		report.AddFinding(Finding{
			Severity: SeverityError,
			Code:     "META_SCHEMA_VALIDATION",
			Path:     path,
			Message:  msg,
			Hint:     "Ensure the schema conforms to the profile meta-schema",
		})
	} else {
		for _, cause := range ve.Causes {
			addValidationErrors(cause, report, path)
		}
	}
}

// instanceLocationToPath converts a JSON pointer segments slice to a path string.
func instanceLocationToPath(loc []string) string {
	if len(loc) == 0 {
		return ""
	}
	return "/" + strings.Join(loc, "/")
}

// yamlToJSON converts YAML-decoded types to JSON-compatible types.
// YAML v3 decodes maps as map[string]interface{} which is fine,
// but some values might need conversion.
func yamlToJSON(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		m := make(map[string]interface{}, len(val))
		for k, v := range val {
			m[k] = yamlToJSON(v)
		}
		return m
	case map[interface{}]interface{}:
		m := make(map[string]interface{}, len(val))
		for k, v := range val {
			m[fmt.Sprintf("%v", k)] = yamlToJSON(v)
		}
		return m
	case []interface{}:
		for i, v := range val {
			val[i] = yamlToJSON(v)
		}
		return val
	default:
		return v
	}
}
