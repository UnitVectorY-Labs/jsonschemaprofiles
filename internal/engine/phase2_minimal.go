package engine

import (
	"fmt"
	"sort"
)

func validatePhase2Minimal(candidateMap map[string]interface{}, rawBytes []byte, report *Report) {
	// Root must be type: object (already enforced by meta-schema, but double check)
	rootType, _ := candidateMap["type"]
	if rootType != "object" {
		report.AddFinding(Finding{
			Severity: SeverityError,
			Code:     "MINIMAL_ROOT_NOT_OBJECT",
			Path:     "",
			Message:  "Root schema must have type: object",
		})
	}

	// Check object discipline at root
	checkMinimalObjectDiscipline(candidateMap, "", report)

	// Check nested object discipline
	checkMinimalNestedDiscipline(candidateMap, "", report)
}

func checkMinimalObjectDiscipline(obj map[string]interface{}, path string, report *Report) {
	// Check additionalProperties is false
	ap, hasAP := obj["additionalProperties"]
	if !hasAP {
		report.AddFinding(Finding{
			Severity: SeverityError,
			Code:     "MINIMAL_MISSING_ADDITIONAL_PROPERTIES",
			Path:     path,
			Message:  "Object schema must have additionalProperties: false",
			Hint:     "Add \"additionalProperties\": false",
		})
	} else if ap != false {
		report.AddFinding(Finding{
			Severity: SeverityError,
			Code:     "MINIMAL_ADDITIONAL_PROPERTIES_NOT_FALSE",
			Path:     path,
			Message:  "additionalProperties must be exactly false",
		})
	}

	// Check required matches properties
	props, hasProps := obj["properties"]
	req, hasReq := obj["required"]

	if hasProps {
		propsMap, _ := props.(map[string]interface{})
		if propsMap != nil {
			propKeys := make(map[string]bool)
			for k := range propsMap {
				propKeys[k] = true
			}

			if !hasReq {
				report.AddFinding(Finding{
					Severity: SeverityError,
					Code:     "MINIMAL_MISSING_REQUIRED",
					Path:     path,
					Message:  "Object with properties must have required listing all property keys",
				})
			} else if reqArr, ok := req.([]interface{}); ok {
				reqKeys := make(map[string]bool)
				for _, r := range reqArr {
					if s, ok := r.(string); ok {
						reqKeys[s] = true
					}
				}
				missing := []string{}
				for k := range propKeys {
					if !reqKeys[k] {
						missing = append(missing, k)
					}
				}
				sort.Strings(missing)
				if len(missing) > 0 {
					report.AddFinding(Finding{
						Severity: SeverityError,
						Code:     "MINIMAL_REQUIRED_MISSING_KEYS",
						Path:     path,
						Message:  fmt.Sprintf("required is missing property keys: %v", missing),
					})
				}
				extra := []string{}
				for k := range reqKeys {
					if !propKeys[k] {
						extra = append(extra, k)
					}
				}
				sort.Strings(extra)
				if len(extra) > 0 {
					report.AddFinding(Finding{
						Severity: SeverityError,
						Code:     "MINIMAL_REQUIRED_EXTRA_KEYS",
						Path:     path,
						Message:  fmt.Sprintf("required contains keys not in properties: %v", extra),
					})
				}
			}
		}
	}
}

func checkMinimalNestedDiscipline(node interface{}, path string, report *Report) {
	obj, ok := node.(map[string]interface{})
	if !ok {
		return
	}

	if props, ok := obj["properties"]; ok {
		if propsMap, ok := props.(map[string]interface{}); ok {
			for name, propSchema := range propsMap {
				propPath := path + "/properties/" + name
				if propObj, ok := propSchema.(map[string]interface{}); ok {
					if isObjectType(propSchema) {
						checkMinimalObjectDiscipline(propObj, propPath, report)
					}
					checkMinimalNestedDiscipline(propSchema, propPath, report)
				}
			}
		}
	}

	if items, ok := obj["items"]; ok {
		if itemObj, ok := items.(map[string]interface{}); ok {
			if isObjectType(items) {
				checkMinimalObjectDiscipline(itemObj, path+"/items", report)
			}
			checkMinimalNestedDiscipline(items, path+"/items", report)
		}
	}
}
