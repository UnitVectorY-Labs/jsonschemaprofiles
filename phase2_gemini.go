package jsonschemaprofiles

import (
	"fmt"
)

func validatePhase2Gemini(candidateMap map[string]interface{}, report *Report, requirePropertyOrdering bool) {
	checkGeminiSchema(candidateMap, "", report, requirePropertyOrdering)
}

func checkGeminiSchema(node interface{}, path string, report *Report, requirePropertyOrdering bool) {
	obj, ok := node.(map[string]interface{})
	if !ok {
		return
	}

	// Check required entries exist in properties
	if props, ok := obj["properties"]; ok {
		if propsMap, ok := props.(map[string]interface{}); ok {
			if req, ok := obj["required"]; ok {
				if reqArr, ok := req.([]interface{}); ok {
					for _, r := range reqArr {
						if s, ok := r.(string); ok {
							if _, exists := propsMap[s]; !exists {
								report.AddFinding(Finding{
									Severity: SeverityError,
									Code:     "GEMINI_REQUIRED_NOT_IN_PROPERTIES",
									Path:     path,
									Message:  fmt.Sprintf("required entry \"%s\" does not exist in properties", s),
									Hint:     "Add the property to properties or remove it from required",
								})
							}
						}
					}
				}
			}

			// Check propertyOrdering if present
			if po, ok := obj["propertyOrdering"]; ok {
				if poArr, ok := po.([]interface{}); ok {
					seen := make(map[string]bool)
					for _, p := range poArr {
						if s, ok := p.(string); ok {
							// Check exists in properties
							if _, exists := propsMap[s]; !exists {
								report.AddFinding(Finding{
									Severity: SeverityError,
									Code:     "GEMINI_PROPERTY_ORDERING_NOT_IN_PROPERTIES",
									Path:     path,
									Message:  fmt.Sprintf("propertyOrdering entry \"%s\" does not exist in properties", s),
									Hint:     "Only include property names that exist in properties",
								})
							}
							// Check duplicates
							if seen[s] {
								report.AddFinding(Finding{
									Severity: SeverityError,
									Code:     "GEMINI_PROPERTY_ORDERING_DUPLICATE",
									Path:     path,
									Message:  fmt.Sprintf("propertyOrdering contains duplicate entry \"%s\"", s),
									Hint:     "Remove duplicate entries from propertyOrdering",
								})
							}
							seen[s] = true
						}
					}
				}
			}

			// For Gemini 2.0: enforce propertyOrdering is present on object schemas
			if requirePropertyOrdering {
				if _, hasPO := obj["propertyOrdering"]; !hasPO {
					report.AddFinding(Finding{
						Severity: SeverityError,
						Code:     "GEMINI_2_0_MISSING_PROPERTY_ORDERING",
						Path:     path,
						Message:  "Object schema must include propertyOrdering for Gemini 2.0 compatibility",
						Hint:     "Add propertyOrdering array listing property names in desired order",
					})
				}
			}

			// Recurse into properties
			for name, propSchema := range propsMap {
				checkGeminiSchema(propSchema, path+"/properties/"+name, report, requirePropertyOrdering)
			}
		}
	}

	// Array semantics: when type includes array, enforce items or prefixItems
	if isArrayTypeObj(obj) {
		_, hasItems := obj["items"]
		_, hasPrefixItems := obj["prefixItems"]
		if !hasItems && !hasPrefixItems {
			report.AddFinding(Finding{
				Severity: SeverityError,
				Code:     "GEMINI_ARRAY_MISSING_ITEMS",
				Path:     path,
				Message:  "Array schema must have either items or prefixItems",
				Hint:     "Add items or prefixItems to the array schema",
			})
		}
	}

	// Recurse into items
	if items, ok := obj["items"]; ok {
		checkGeminiSchema(items, path+"/items", report, requirePropertyOrdering)
	}

	// Recurse into prefixItems
	if prefixItems, ok := obj["prefixItems"]; ok {
		if arr, ok := prefixItems.([]interface{}); ok {
			for i, item := range arr {
				checkGeminiSchema(item, fmt.Sprintf("%s/prefixItems/%d", path, i), report, requirePropertyOrdering)
			}
		}
	}

	// Recurse into additionalProperties if it's a schema
	if ap, ok := obj["additionalProperties"]; ok {
		if apMap, ok := ap.(map[string]interface{}); ok {
			checkGeminiSchema(apMap, path+"/additionalProperties", report, requirePropertyOrdering)
		}
	}
}

// isArrayTypeObj checks if the object has type: array or type: ["array", ...]
func isArrayTypeObj(obj map[string]interface{}) bool {
	t, ok := obj["type"]
	if !ok {
		return false
	}
	switch tv := t.(type) {
	case string:
		return tv == "array"
	case []interface{}:
		for _, v := range tv {
			if s, ok := v.(string); ok && s == "array" {
				return true
			}
		}
	}
	return false
}

// isObjectTypeObj checks if the object has type: object or type: ["object", ...]
func isObjectTypeObj(obj map[string]interface{}) bool {
	t, ok := obj["type"]
	if !ok {
		return false
	}
	switch tv := t.(type) {
	case string:
		return tv == "object"
	case []interface{}:
		for _, v := range tv {
			if s, ok := v.(string); ok && s == "object" {
				return true
			}
		}
	}
	return false
}
