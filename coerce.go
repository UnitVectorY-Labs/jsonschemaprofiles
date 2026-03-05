package jsonschemaprofiles

import "github.com/UnitVectorY-Labs/jsonschemaprofiles/internal/engine"

// CoerceMode specifies the coercion aggressiveness.
type CoerceMode = engine.CoerceMode

const (
	CoerceModeConservative CoerceMode = engine.CoerceModeConservative
	CoerceModePermissive   CoerceMode = engine.CoerceModePermissive
	CoerceModeOff          CoerceMode = engine.CoerceModeOff
)

// CoerceOptions configures coercion behavior.
type CoerceOptions = engine.CoerceOptions

// CoerceSchema attempts to make a schema compliant with a profile.
// Returns the coerced schema bytes, a report of changes, and whether any changes were made.
func CoerceSchema(profileID ProfileID, schemaBytes []byte, opts *CoerceOptions) ([]byte, *Report, bool, error) {
	info, err := GetProfileInfo(profileID)
	if err != nil {
		return nil, nil, false, err
	}

	metaSchemaYAML, err := GetProfileSchema(profileID)
	if err != nil {
		return nil, nil, false, err
	}

	return engine.CoerceSchema(string(profileID), info.SchemaFile, metaSchemaYAML, schemaBytes, opts)
}
