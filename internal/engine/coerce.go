package engine

import (
	"encoding/json"
	"fmt"
	"sort"
)

// CoerceMode specifies the coercion aggressiveness.
type CoerceMode string

const (
	CoerceModeConservative CoerceMode = "conservative"
	CoerceModePermissive   CoerceMode = "permissive"
	CoerceModeOff          CoerceMode = "off"
)

// CoerceOptions configures coercion behavior.
type CoerceOptions struct {
	Mode   CoerceMode
	DryRun bool
}

// CoerceSchema attempts to make a schema compliant with a profile.
// Returns the coerced schema bytes, a report of changes, and whether any changes were made.
func CoerceSchema(profileID string, schemaFile string, metaSchemaYAML []byte, schemaBytes []byte, opts *CoerceOptions) ([]byte, *Report, bool, error) {
	if opts == nil {
		opts = &CoerceOptions{Mode: CoerceModeConservative}
	}

	if opts.Mode == CoerceModeOff {
		report, err := ValidateSchema(profileID, schemaFile, metaSchemaYAML, schemaBytes, nil)
		return schemaBytes, report, false, err
	}

	report := NewReport()

	// Parse into a modifiable structure
	var schema map[string]interface{}
	if err := json.Unmarshal(schemaBytes, &schema); err != nil {
		return nil, nil, false, fmt.Errorf("failed to parse schema: %w", err)
	}

	changed := false

	switch profileID {
	case profileOpenAI:
		changed = coerceOpenAI(schema, "", report, opts)
	case profileGemini:
		changed = coerceGemini(schema, "", report, opts, false)
	case profileGemini20:
		changed = coerceGemini(schema, "", report, opts, true)
	case profileMinimal:
		changed = coerceMinimal(schema, "", report, opts)
	default:
		return nil, nil, false, fmt.Errorf("unknown profile: %s", profileID)
	}

	// Validate the post-coercion schema to ensure the resulting report reflects
	// actual compliance, not only applied coercion actions.
	validationTarget := schemaBytes
	if changed {
		var err error
		validationTarget, err = json.Marshal(schema)
		if err != nil {
			return nil, nil, false, fmt.Errorf("failed to serialize coerced schema for validation: %w", err)
		}
	}
	validationReport, err := ValidateSchema(profileID, schemaFile, metaSchemaYAML, validationTarget, nil)
	if err != nil {
		return nil, nil, false, fmt.Errorf("failed to validate coerced schema: %w", err)
	}
	mergeReports(report, validationReport)

	if opts.DryRun {
		// Return original bytes, but the report reflects whether the
		// proposed/coerced result would be compliant.
		return schemaBytes, report, changed, nil
	}

	// Keep exact input bytes when no semantic changes were applied.
	if !changed {
		return schemaBytes, report, false, nil
	}

	// Serialize the coerced schema
	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return nil, nil, false, fmt.Errorf("failed to serialize coerced schema: %w", err)
	}

	return out, report, changed, nil
}

// coerceOpenAI applies OpenAI-specific coercions.
func coerceOpenAI(obj map[string]interface{}, path string, report *Report, opts *CoerceOptions) bool {
	changed := false

	// Add type: object if properties present and type missing
	if _, hasType := obj["type"]; !hasType {
		if _, hasProps := obj["properties"]; hasProps {
			obj["type"] = "object"
			changed = true
			report.AddFinding(Finding{
				Severity: SeverityInfo,
				Code:     "COERCE_ADD_TYPE_OBJECT",
				Path:     path,
				Message:  "Added type: object (inferred from properties)",
				Rule:     "add-missing-type",
			})
		} else if _, hasItems := obj["items"]; hasItems {
			obj["type"] = "array"
			changed = true
			report.AddFinding(Finding{
				Severity: SeverityInfo,
				Code:     "COERCE_ADD_TYPE_ARRAY",
				Path:     path,
				Message:  "Added type: array (inferred from items)",
				Rule:     "add-missing-type",
			})
		}
	}

	// Add additionalProperties: false if missing and type is object
	if isObjectType(obj) {
		if _, hasAP := obj["additionalProperties"]; !hasAP {
			obj["additionalProperties"] = false
			changed = true
			report.AddFinding(Finding{
				Severity: SeverityInfo,
				Code:     "COERCE_ADD_ADDITIONAL_PROPERTIES_FALSE",
				Path:     path,
				Message:  "Added additionalProperties: false",
				Rule:     "add-additional-properties",
			})
		}

		// Fill required if missing
		if props, ok := obj["properties"]; ok {
			if propsMap, ok := props.(map[string]interface{}); ok {
				if _, hasReq := obj["required"]; !hasReq {
					keys := make([]string, 0, len(propsMap))
					for k := range propsMap {
						keys = append(keys, k)
					}
					// Sort for deterministic output
					sortStrings(keys)
					reqArr := make([]interface{}, len(keys))
					for i, k := range keys {
						reqArr[i] = k
					}
					obj["required"] = reqArr
					changed = true
					report.AddFinding(Finding{
						Severity: SeverityInfo,
						Code:     "COERCE_ADD_REQUIRED",
						Path:     path,
						Message:  "Added required with all property keys",
						Rule:     "add-required",
					})
				} else {
					// Add missing keys to required
					if reqArr, ok := obj["required"].([]interface{}); ok {
						reqKeys := make(map[string]bool)
						for _, r := range reqArr {
							if s, ok := r.(string); ok {
								reqKeys[s] = true
							}
						}
						added := false
						for k := range propsMap {
							if !reqKeys[k] {
								reqArr = append(reqArr, k)
								added = true
								report.AddFinding(Finding{
									Severity: SeverityWarning,
									Code:     "COERCE_ADD_REQUIRED_KEY",
									Path:     path,
									Message:  fmt.Sprintf("Added missing key \"%s\" to required", k),
									Rule:     "add-required-key",
								})
							}
						}
						if added {
							obj["required"] = reqArr
							changed = true
						}
					}
				}
			}
		}
	}

	// Recurse into nested schemas
	if props, ok := obj["properties"]; ok {
		if propsMap, ok := props.(map[string]interface{}); ok {
			for name, propSchema := range propsMap {
				if propObj, ok := propSchema.(map[string]interface{}); ok {
					if coerceOpenAI(propObj, path+"/properties/"+name, report, opts) {
						changed = true
					}
				}
			}
		}
	}
	if items, ok := obj["items"]; ok {
		if itemsObj, ok := items.(map[string]interface{}); ok {
			if coerceOpenAI(itemsObj, path+"/items", report, opts) {
				changed = true
			}
		}
	}
	if anyOf, ok := obj["anyOf"]; ok {
		if arr, ok := anyOf.([]interface{}); ok {
			for i, item := range arr {
				if itemObj, ok := item.(map[string]interface{}); ok {
					if coerceOpenAI(itemObj, fmt.Sprintf("%s/anyOf/%d", path, i), report, opts) {
						changed = true
					}
				}
			}
		}
	}
	if defs, ok := obj["$defs"]; ok {
		if defsMap, ok := defs.(map[string]interface{}); ok {
			for name, defSchema := range defsMap {
				if defObj, ok := defSchema.(map[string]interface{}); ok {
					if coerceOpenAI(defObj, path+"/$defs/"+name, report, opts) {
						changed = true
					}
				}
			}
		}
	}

	// Permissive mode: drop unsupported keywords
	if opts.Mode == CoerceModePermissive {
		unsupported := []string{"allOf", "not", "dependentRequired", "dependentSchemas", "if", "then", "else"}
		for _, kw := range unsupported {
			if _, has := obj[kw]; has {
				delete(obj, kw)
				changed = true
				report.AddFinding(Finding{
					Severity: SeverityWarning,
					Code:     "COERCE_DROP_UNSUPPORTED_KEYWORD",
					Path:     path,
					Message:  fmt.Sprintf("Dropped unsupported keyword \"%s\"", kw),
					Rule:     "drop-unsupported",
				})
			}
		}
	}

	return changed
}

// coerceGemini applies Gemini-specific coercions.
func coerceGemini(obj map[string]interface{}, path string, report *Report, opts *CoerceOptions, requirePropertyOrdering bool) bool {
	changed := false

	// Add type if missing and inferable
	if _, hasType := obj["type"]; !hasType {
		if _, hasProps := obj["properties"]; hasProps {
			obj["type"] = "object"
			changed = true
			report.AddFinding(Finding{
				Severity: SeverityInfo,
				Code:     "COERCE_ADD_TYPE_OBJECT",
				Path:     path,
				Message:  "Added type: object (inferred from properties)",
				Rule:     "add-missing-type",
			})
		} else if _, hasItems := obj["items"]; hasItems {
			obj["type"] = "array"
			changed = true
			report.AddFinding(Finding{
				Severity: SeverityInfo,
				Code:     "COERCE_ADD_TYPE_ARRAY",
				Path:     path,
				Message:  "Added type: array (inferred from items)",
				Rule:     "add-missing-type",
			})
		} else if _, hasPrefixItems := obj["prefixItems"]; hasPrefixItems {
			obj["type"] = "array"
			changed = true
			report.AddFinding(Finding{
				Severity: SeverityInfo,
				Code:     "COERCE_ADD_TYPE_ARRAY",
				Path:     path,
				Message:  "Added type: array (inferred from prefixItems)",
				Rule:     "add-missing-type",
			})
		}
	}

	// For Gemini 2.0: add propertyOrdering if missing and properties present
	if requirePropertyOrdering && isObjectTypeObj(obj) {
		if _, hasPO := obj["propertyOrdering"]; !hasPO {
			if props, ok := obj["properties"]; ok {
				if propsMap, ok := props.(map[string]interface{}); ok {
					keys := make([]string, 0, len(propsMap))
					for k := range propsMap {
						keys = append(keys, k)
					}
					sortStrings(keys)
					poArr := make([]interface{}, len(keys))
					for i, k := range keys {
						poArr[i] = k
					}
					obj["propertyOrdering"] = poArr
					changed = true
					report.AddFinding(Finding{
						Severity: SeverityWarning,
						Code:     "COERCE_ADD_PROPERTY_ORDERING",
						Path:     path,
						Message:  "Added propertyOrdering derived from properties keys (ordering may not match source)",
						Rule:     "add-property-ordering",
					})
				}
			}
		}
	}

	// Recurse
	if props, ok := obj["properties"]; ok {
		if propsMap, ok := props.(map[string]interface{}); ok {
			for name, propSchema := range propsMap {
				if propObj, ok := propSchema.(map[string]interface{}); ok {
					if coerceGemini(propObj, path+"/properties/"+name, report, opts, requirePropertyOrdering) {
						changed = true
					}
				}
			}
		}
	}
	if items, ok := obj["items"]; ok {
		if itemsObj, ok := items.(map[string]interface{}); ok {
			if coerceGemini(itemsObj, path+"/items", report, opts, requirePropertyOrdering) {
				changed = true
			}
		}
	}
	if prefixItems, ok := obj["prefixItems"]; ok {
		if arr, ok := prefixItems.([]interface{}); ok {
			for i, item := range arr {
				if itemObj, ok := item.(map[string]interface{}); ok {
					if coerceGemini(itemObj, fmt.Sprintf("%s/prefixItems/%d", path, i), report, opts, requirePropertyOrdering) {
						changed = true
					}
				}
			}
		}
	}
	if ap, ok := obj["additionalProperties"]; ok {
		if apMap, ok := ap.(map[string]interface{}); ok {
			if coerceGemini(apMap, path+"/additionalProperties", report, opts, requirePropertyOrdering) {
				changed = true
			}
		}
	}

	return changed
}

// coerceMinimal applies Minimal profile coercions (OpenAI-like discipline).
func coerceMinimal(obj map[string]interface{}, path string, report *Report, opts *CoerceOptions) bool {
	changed := false

	// Add type if missing and inferable
	if _, hasType := obj["type"]; !hasType {
		if _, hasProps := obj["properties"]; hasProps {
			obj["type"] = "object"
			changed = true
			report.AddFinding(Finding{
				Severity: SeverityInfo,
				Code:     "COERCE_ADD_TYPE_OBJECT",
				Path:     path,
				Message:  "Added type: object (inferred from properties)",
				Rule:     "add-missing-type",
			})
		} else if _, hasItems := obj["items"]; hasItems {
			obj["type"] = "array"
			changed = true
			report.AddFinding(Finding{
				Severity: SeverityInfo,
				Code:     "COERCE_ADD_TYPE_ARRAY",
				Path:     path,
				Message:  "Added type: array (inferred from items)",
				Rule:     "add-missing-type",
			})
		}
	}

	// Add additionalProperties: false if missing and type is object
	if isObjectType(obj) {
		if _, hasAP := obj["additionalProperties"]; !hasAP {
			obj["additionalProperties"] = false
			changed = true
			report.AddFinding(Finding{
				Severity: SeverityInfo,
				Code:     "COERCE_ADD_ADDITIONAL_PROPERTIES_FALSE",
				Path:     path,
				Message:  "Added additionalProperties: false",
				Rule:     "add-additional-properties",
			})
		}

		if props, ok := obj["properties"]; ok {
			if propsMap, ok := props.(map[string]interface{}); ok {
				if _, hasReq := obj["required"]; !hasReq {
					keys := make([]string, 0, len(propsMap))
					for k := range propsMap {
						keys = append(keys, k)
					}
					sortStrings(keys)
					reqArr := make([]interface{}, len(keys))
					for i, k := range keys {
						reqArr[i] = k
					}
					obj["required"] = reqArr
					changed = true
					report.AddFinding(Finding{
						Severity: SeverityInfo,
						Code:     "COERCE_ADD_REQUIRED",
						Path:     path,
						Message:  "Added required with all property keys",
						Rule:     "add-required",
					})
				} else {
					if reqArr, ok := obj["required"].([]interface{}); ok {
						reqKeys := make(map[string]bool)
						for _, r := range reqArr {
							if s, ok := r.(string); ok {
								reqKeys[s] = true
							}
						}
						added := false
						for k := range propsMap {
							if !reqKeys[k] {
								reqArr = append(reqArr, k)
								added = true
								report.AddFinding(Finding{
									Severity: SeverityWarning,
									Code:     "COERCE_ADD_REQUIRED_KEY",
									Path:     path,
									Message:  fmt.Sprintf("Added missing key \"%s\" to required", k),
									Rule:     "add-required-key",
								})
							}
						}
						if added {
							obj["required"] = reqArr
							changed = true
						}
					}
				}
			}
		}
	}

	// Recurse
	if props, ok := obj["properties"]; ok {
		if propsMap, ok := props.(map[string]interface{}); ok {
			for name, propSchema := range propsMap {
				if propObj, ok := propSchema.(map[string]interface{}); ok {
					if coerceMinimal(propObj, path+"/properties/"+name, report, opts) {
						changed = true
					}
				}
			}
		}
	}
	if items, ok := obj["items"]; ok {
		if itemsObj, ok := items.(map[string]interface{}); ok {
			if coerceMinimal(itemsObj, path+"/items", report, opts) {
				changed = true
			}
		}
	}

	return changed
}

func sortStrings(s []string) {
	sort.Strings(s)
}

func mergeReports(dst *Report, src *Report) {
	if dst == nil || src == nil {
		return
	}
	for _, finding := range src.Findings {
		dst.AddFinding(finding)
	}
}
