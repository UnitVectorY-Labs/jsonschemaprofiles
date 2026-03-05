package jsonschemaprofiles

import "github.com/UnitVectorY-Labs/jsonschemaprofiles/internal/engine"

// ValidateOptions configures validation behavior.
type ValidateOptions = engine.ValidateOptions

// ValidateSchema validates a candidate JSON Schema against a profile.
// It runs Phase 1 (meta-schema validation) and Phase 2 (provider-specific code checks).
func ValidateSchema(profileID ProfileID, schemaBytes []byte, opts *ValidateOptions) (*Report, error) {
	info, err := GetProfileInfo(profileID)
	if err != nil {
		return nil, err
	}

	metaSchemaYAML, err := GetProfileSchema(profileID)
	if err != nil {
		return nil, err
	}

	return engine.ValidateSchema(string(profileID), info.SchemaFile, metaSchemaYAML, schemaBytes, opts)
}
