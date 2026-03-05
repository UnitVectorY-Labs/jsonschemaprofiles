package jsonschemaprofiles

import (
	"fmt"
	"strings"
)

// SchemaMetrics holds computed metrics from a schema graph traversal.
type SchemaMetrics struct {
	TotalProperties int
	MaxDepth        int
	TotalEnumValues int
	StringBudget    int
	LargeEnums      []largeEnum // enums with >250 values
}

type largeEnum struct {
	Path         string
	ValueCount   int
	StringLength int
}

const maxTraversalBudget = 100000 // Maximum nodes to visit to prevent DoS

// traverseContext holds state for a single traversal.
type traverseContext struct {
	defs      map[string]interface{}
	visited   map[string]bool // for cycle detection on $ref
	metrics   SchemaMetrics
	nodeCount int
}

func newTraverseContext(candidateMap map[string]interface{}) *traverseContext {
	ctx := &traverseContext{
		visited: make(map[string]bool),
	}
	// Extract $defs if present
	if defs, ok := candidateMap["$defs"]; ok {
		if defsMap, ok := defs.(map[string]interface{}); ok {
			ctx.defs = defsMap
			// Count string budget from $defs names
			for name := range defsMap {
				ctx.metrics.StringBudget += len(name)
			}
		}
	}
	return ctx
}

// traverseSchema walks the schema graph, computing metrics and following $refs.
func (ctx *traverseContext) traverseSchema(node interface{}, path string, depth int) {
	ctx.nodeCount++
	if ctx.nodeCount > maxTraversalBudget {
		return
	}

	obj, ok := node.(map[string]interface{})
	if !ok {
		return
	}

	// Handle $ref
	if ref, ok := obj["$ref"]; ok {
		if refStr, ok := ref.(string); ok {
			ctx.resolveAndTraverse(refStr, path, depth)
			return
		}
	}

	// Track depth
	if depth > ctx.metrics.MaxDepth {
		ctx.metrics.MaxDepth = depth
	}

	// Count properties and string budget
	if props, ok := obj["properties"]; ok {
		if propsMap, ok := props.(map[string]interface{}); ok {
			ctx.metrics.TotalProperties += len(propsMap)
			for name, propSchema := range propsMap {
				ctx.metrics.StringBudget += len(name)
				nextDepth := depth
				// If this property is an object type, increase depth
				if isObjectType(propSchema) {
					nextDepth = depth + 1
				}
				ctx.traverseSchema(propSchema, path+"/properties/"+name, nextDepth)
			}
		}
	}

	// Count enum values and string budget
	if enumVal, ok := obj["enum"]; ok {
		if enumArr, ok := enumVal.([]interface{}); ok {
			ctx.metrics.TotalEnumValues += len(enumArr)
			totalStr := 0
			for _, v := range enumArr {
				if s, ok := v.(string); ok {
					ctx.metrics.StringBudget += len(s)
					totalStr += len(s)
				}
			}
			if len(enumArr) > 250 {
				ctx.metrics.LargeEnums = append(ctx.metrics.LargeEnums, largeEnum{
					Path:         path,
					ValueCount:   len(enumArr),
					StringLength: totalStr,
				})
			}
		}
	}

	// Count const values string budget
	if constVal, ok := obj["const"]; ok {
		if s, ok := constVal.(string); ok {
			ctx.metrics.StringBudget += len(s)
		}
	}

	// Traverse items
	if items, ok := obj["items"]; ok {
		ctx.traverseSchema(items, path+"/items", depth)
	}

	// Traverse prefixItems
	if prefixItems, ok := obj["prefixItems"]; ok {
		if arr, ok := prefixItems.([]interface{}); ok {
			for i, item := range arr {
				ctx.traverseSchema(item, fmt.Sprintf("%s/prefixItems/%d", path, i), depth)
			}
		}
	}

	// Traverse anyOf
	if anyOf, ok := obj["anyOf"]; ok {
		if arr, ok := anyOf.([]interface{}); ok {
			for i, item := range arr {
				ctx.traverseSchema(item, fmt.Sprintf("%s/anyOf/%d", path, i), depth)
			}
		}
	}

	// Traverse $defs (for completeness, but they are traversed via $ref)
	if defs, ok := obj["$defs"]; ok {
		if defsMap, ok := defs.(map[string]interface{}); ok {
			for name, defSchema := range defsMap {
				ctx.traverseSchema(defSchema, path+"/$defs/"+name, depth)
			}
		}
	}

	// Traverse additionalProperties if it's a schema (not boolean)
	if ap, ok := obj["additionalProperties"]; ok {
		if apMap, ok := ap.(map[string]interface{}); ok {
			ctx.traverseSchema(apMap, path+"/additionalProperties", depth)
		}
	}
}

// resolveAndTraverse follows a $ref and traverses the target.
func (ctx *traverseContext) resolveAndTraverse(ref string, path string, depth int) {
	if !strings.HasPrefix(ref, "#") {
		return // Only support local refs
	}

	// Cycle detection
	if ctx.visited[ref] {
		return
	}
	ctx.visited[ref] = true
	defer func() { ctx.visited[ref] = false }()

	// Parse the JSON pointer
	pointer := strings.TrimPrefix(ref, "#")
	if pointer == "" || pointer == "/" {
		return
	}

	parts := strings.Split(strings.TrimPrefix(pointer, "/"), "/")
	if len(parts) < 2 || parts[0] != "$defs" {
		return
	}

	defName := parts[1]
	if ctx.defs == nil {
		return
	}

	target, ok := ctx.defs[defName]
	if !ok {
		return
	}

	ctx.traverseSchema(target, path, depth)
}

// isObjectType returns true if the schema node declares type: object or type: ["object", ...].
func isObjectType(node interface{}) bool {
	obj, ok := node.(map[string]interface{})
	if !ok {
		return false
	}
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

// isArrayType returns true if the schema node declares type: array or type: ["array", ...].
func isArrayType(node interface{}) bool {
	obj, ok := node.(map[string]interface{})
	if !ok {
		return false
	}
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
