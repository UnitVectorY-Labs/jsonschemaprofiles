package jsonschemaprofiles

import (
	"fmt"
	"sort"
)

// OpenAI limits
const (
	openAIMaxProperties   = 5000
	openAIMaxDepth        = 10
	openAIMaxStringBudget = 120000
	openAIMaxEnumTotal    = 1000
	openAILargeEnumValues = 250
	openAILargeEnumStrLen = 15000
)

// OpenAI fine-tuned model unsupported keywords
var openAIFineTunedUnsupported = map[string][]string{
	"string": {"minLength", "maxLength", "pattern", "format"},
	"number": {"minimum", "maximum", "multipleOf"},
	"object": {"patternProperties", "unevaluatedProperties", "propertyNames", "minProperties", "maxProperties"},
	"array":  {"unevaluatedItems", "contains", "minContains", "maxContains", "minItems", "maxItems", "uniqueItems"},
}

func validatePhase2OpenAI(candidateMap map[string]interface{}, rawBytes []byte, report *Report, opts *ValidateOptions) {
	// Check root is type: object
	rootType, _ := candidateMap["type"]
	if rootType != "object" {
		report.AddFinding(Finding{
			Severity: SeverityError,
			Code:     "OPENAI_ROOT_NOT_OBJECT",
			Path:     "",
			Message:  "Root schema must have type: object",
			Hint:     "Set the root schema type to \"object\"",
		})
	}

	// Check root is not anyOf
	if _, hasAnyOf := candidateMap["anyOf"]; hasAnyOf {
		report.AddFinding(Finding{
			Severity: SeverityError,
			Code:     "OPENAI_ROOT_ANYOF",
			Path:     "",
			Message:  "Root schema must not use anyOf",
			Hint:     "Remove anyOf from the root schema",
		})
	}

	// Check object discipline at root
	checkOpenAIObjectDiscipline(candidateMap, "", report)

	// Traverse for nested checks and metrics
	ctx := newTraverseContext(candidateMap)
	ctx.traverseSchema(candidateMap, "", 0)

	// Check nested object discipline
	checkOpenAINestedObjectDiscipline(candidateMap, "", report, ctx, make(map[string]bool))

	// Check limits
	if ctx.metrics.TotalProperties > openAIMaxProperties {
		report.AddFinding(Finding{
			Severity: SeverityError,
			Code:     "OPENAI_BUDGET_PROPERTIES_EXCEEDED",
			Path:     "",
			Message:  fmt.Sprintf("Total object properties (%d) exceeds limit of %d", ctx.metrics.TotalProperties, openAIMaxProperties),
		})
	}

	if ctx.metrics.MaxDepth > openAIMaxDepth {
		report.AddFinding(Finding{
			Severity: SeverityError,
			Code:     "OPENAI_BUDGET_DEPTH_EXCEEDED",
			Path:     "",
			Message:  fmt.Sprintf("Maximum object nesting depth (%d) exceeds limit of %d", ctx.metrics.MaxDepth, openAIMaxDepth),
		})
	}

	if ctx.metrics.StringBudget > openAIMaxStringBudget {
		report.AddFinding(Finding{
			Severity: SeverityError,
			Code:     "OPENAI_BUDGET_STRING_EXCEEDED",
			Path:     "",
			Message:  fmt.Sprintf("Total string budget (%d) exceeds limit of %d characters", ctx.metrics.StringBudget, openAIMaxStringBudget),
		})
	}

	if ctx.metrics.TotalEnumValues > openAIMaxEnumTotal {
		report.AddFinding(Finding{
			Severity: SeverityError,
			Code:     "OPENAI_BUDGET_ENUM_TOTAL_EXCEEDED",
			Path:     "",
			Message:  fmt.Sprintf("Total enum values (%d) exceeds limit of %d", ctx.metrics.TotalEnumValues, openAIMaxEnumTotal),
		})
	}

	for _, le := range ctx.metrics.LargeEnums {
		if le.StringLength > openAILargeEnumStrLen {
			report.AddFinding(Finding{
				Severity: SeverityError,
				Code:     "OPENAI_BUDGET_LARGE_ENUM_STRING_EXCEEDED",
				Path:     le.Path,
				Message:  fmt.Sprintf("Enum with %d values has combined string length %d exceeding limit of %d", le.ValueCount, le.StringLength, openAILargeEnumStrLen),
			})
		}
	}

	// Fine-tuned model checks
	if opts != nil && opts.ModelTarget == "fine-tuned" {
		checkOpenAIFineTunedKeywords(candidateMap, "", report)
	}
}

// checkOpenAIObjectDiscipline checks that an object schema has additionalProperties: false
// and that required matches properties keys exactly.
func checkOpenAIObjectDiscipline(obj map[string]interface{}, path string, report *Report) {
	// Check additionalProperties
	ap, hasAP := obj["additionalProperties"]
	if !hasAP {
		report.AddFinding(Finding{
			Severity: SeverityError,
			Code:     "OPENAI_MISSING_ADDITIONAL_PROPERTIES",
			Path:     path,
			Message:  "Object schema must have additionalProperties: false",
			Hint:     "Add \"additionalProperties\": false",
		})
	} else if ap != false {
		report.AddFinding(Finding{
			Severity: SeverityError,
			Code:     "OPENAI_ADDITIONAL_PROPERTIES_NOT_FALSE",
			Path:     path,
			Message:  "additionalProperties must be exactly false",
			Hint:     "Set \"additionalProperties\": false",
		})
	}

	// Check required matches properties keys
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
					Code:     "OPENAI_MISSING_REQUIRED",
					Path:     path,
					Message:  "Object with properties must have required listing all property keys",
					Hint:     "Add a \"required\" array listing all property names",
				})
			} else if reqArr, ok := req.([]interface{}); ok {
				reqKeys := make(map[string]bool)
				for _, r := range reqArr {
					if s, ok := r.(string); ok {
						reqKeys[s] = true
					}
				}
				// Check for missing required keys
				missingRequired := []string{}
				for k := range propKeys {
					if !reqKeys[k] {
						missingRequired = append(missingRequired, k)
					}
				}
				sort.Strings(missingRequired)
				if len(missingRequired) > 0 {
					report.AddFinding(Finding{
						Severity: SeverityError,
						Code:     "OPENAI_REQUIRED_MISSING_KEYS",
						Path:     path,
						Message:  fmt.Sprintf("required is missing property keys: %v", missingRequired),
						Hint:     "Add the missing keys to the required array",
					})
				}
				// Check for extra required keys
				extraRequired := []string{}
				for k := range reqKeys {
					if !propKeys[k] {
						extraRequired = append(extraRequired, k)
					}
				}
				sort.Strings(extraRequired)
				if len(extraRequired) > 0 {
					report.AddFinding(Finding{
						Severity: SeverityError,
						Code:     "OPENAI_REQUIRED_EXTRA_KEYS",
						Path:     path,
						Message:  fmt.Sprintf("required contains keys not in properties: %v", extraRequired),
						Hint:     "Remove extra keys from required or add them to properties",
					})
				}
			}
		}
	}
}

// checkOpenAINestedObjectDiscipline walks nested object schemas and checks discipline.
func checkOpenAINestedObjectDiscipline(node interface{}, path string, report *Report, ctx *traverseContext, visitedRefs map[string]bool) {
	obj, ok := node.(map[string]interface{})
	if !ok {
		return
	}

	// Handle $ref - resolve and check with cycle detection
	if ref, ok := obj["$ref"]; ok {
		if refStr, ok := ref.(string); ok {
			if visitedRefs[refStr] {
				return
			}
			visitedRefs[refStr] = true
			resolved := resolveRef(refStr, ctx.defs)
			if resolved != nil {
				if resolvedObj, ok := resolved.(map[string]interface{}); ok {
					if isObjectType(resolved) {
						checkOpenAIObjectDiscipline(resolvedObj, path, report)
					}
					checkOpenAINestedObjectDiscipline(resolved, path, report, ctx, visitedRefs)
				}
			}
			delete(visitedRefs, refStr)
			return
		}
	}

	// Check properties
	if props, ok := obj["properties"]; ok {
		if propsMap, ok := props.(map[string]interface{}); ok {
			for name, propSchema := range propsMap {
				propPath := path + "/properties/" + name
				if propObj, ok := propSchema.(map[string]interface{}); ok {
					if isObjectType(propSchema) {
						checkOpenAIObjectDiscipline(propObj, propPath, report)
					}
					checkOpenAINestedObjectDiscipline(propSchema, propPath, report, ctx, visitedRefs)
				}
			}
		}
	}

	// Check items
	if items, ok := obj["items"]; ok {
		if itemObj, ok := items.(map[string]interface{}); ok {
			if isObjectType(items) {
				checkOpenAIObjectDiscipline(itemObj, path+"/items", report)
			}
			checkOpenAINestedObjectDiscipline(items, path+"/items", report, ctx, visitedRefs)
		}
	}

	// Check anyOf
	if anyOf, ok := obj["anyOf"]; ok {
		if arr, ok := anyOf.([]interface{}); ok {
			for i, item := range arr {
				itemPath := fmt.Sprintf("%s/anyOf/%d", path, i)
				if itemObj, ok := item.(map[string]interface{}); ok {
					if isObjectType(item) {
						checkOpenAIObjectDiscipline(itemObj, itemPath, report)
					}
					checkOpenAINestedObjectDiscipline(item, itemPath, report, ctx, visitedRefs)
				}
			}
		}
	}

	// Check $defs
	if defs, ok := obj["$defs"]; ok {
		if defsMap, ok := defs.(map[string]interface{}); ok {
			for name, defSchema := range defsMap {
				defPath := path + "/$defs/" + name
				if defObj, ok := defSchema.(map[string]interface{}); ok {
					if isObjectType(defSchema) {
						checkOpenAIObjectDiscipline(defObj, defPath, report)
					}
					checkOpenAINestedObjectDiscipline(defSchema, defPath, report, ctx, visitedRefs)
				}
			}
		}
	}
}

// resolveRef resolves a local $ref to its target node.
func resolveRef(ref string, defs map[string]interface{}) interface{} {
	if defs == nil {
		return nil
	}
	// Only handle #/$defs/Name pattern
	if len(ref) > 8 && ref[:8] == "#/$defs/" {
		name := ref[8:]
		if target, ok := defs[name]; ok {
			return target
		}
	}
	return nil
}

// checkOpenAIFineTunedKeywords checks for keywords unsupported by fine-tuned models.
func checkOpenAIFineTunedKeywords(node interface{}, path string, report *Report) {
	obj, ok := node.(map[string]interface{})
	if !ok {
		return
	}

	// Determine the type of this schema node
	schemaType := getSchemaType(obj)

	// Check for unsupported keywords based on type
	for typeGroup, keywords := range openAIFineTunedUnsupported {
		if schemaType == typeGroup || typeGroup == "" {
			for _, kw := range keywords {
				if _, has := obj[kw]; has {
					report.AddFinding(Finding{
						Severity: SeverityError,
						Code:     "OPENAI_FINE_TUNED_UNSUPPORTED_KEYWORD",
						Path:     path,
						Message:  fmt.Sprintf("Keyword \"%s\" is not supported for fine-tuned models (type: %s)", kw, typeGroup),
						Hint:     fmt.Sprintf("Remove \"%s\" when targeting fine-tuned models", kw),
					})
				}
			}
		}
	}

	// Recurse into nested schemas
	if props, ok := obj["properties"]; ok {
		if propsMap, ok := props.(map[string]interface{}); ok {
			for name, propSchema := range propsMap {
				checkOpenAIFineTunedKeywords(propSchema, path+"/properties/"+name, report)
			}
		}
	}
	if items, ok := obj["items"]; ok {
		checkOpenAIFineTunedKeywords(items, path+"/items", report)
	}
	if anyOf, ok := obj["anyOf"]; ok {
		if arr, ok := anyOf.([]interface{}); ok {
			for i, item := range arr {
				checkOpenAIFineTunedKeywords(item, fmt.Sprintf("%s/anyOf/%d", path, i), report)
			}
		}
	}
	if defs, ok := obj["$defs"]; ok {
		if defsMap, ok := defs.(map[string]interface{}); ok {
			for name, defSchema := range defsMap {
				checkOpenAIFineTunedKeywords(defSchema, path+"/$defs/"+name, report)
			}
		}
	}
}

// getSchemaType returns the base type string from a schema node.
func getSchemaType(obj map[string]interface{}) string {
	t, ok := obj["type"]
	if !ok {
		return ""
	}
	switch tv := t.(type) {
	case string:
		return tv
	case []interface{}:
		// Return the non-null type
		for _, v := range tv {
			if s, ok := v.(string); ok && s != "null" {
				return s
			}
		}
	}
	return ""
}
