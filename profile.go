package jsonschemaprofiles

import (
	"embed"
	"fmt"
)

//go:embed schemas/*.yaml
var schemaFS embed.FS

// ProfileID identifies a provider profile.
type ProfileID string

const (
	OPENAI_202602  ProfileID = "OPENAI_202602"
	GEMINI_202602  ProfileID = "GEMINI_202602"
	GEMINI_202503  ProfileID = "GEMINI_202503"
	MINIMAL_202602 ProfileID = "MINIMAL_202602"
)

// ProfileInfo describes a profile.
type ProfileInfo struct {
	ID          ProfileID
	Title       string
	Description string
	Baseline    string
	SchemaFile  string // path relative to embed FS root
}

var profiles = map[ProfileID]ProfileInfo{
	OPENAI_202602: {
		ID:          OPENAI_202602,
		Title:       "OpenAI Structured Outputs Profile",
		Description: "Meta-schema for JSON Schemas intended for OpenAI Structured Outputs",
		Baseline:    "202602",
		SchemaFile:  "schemas/openai_202602.yaml",
	},
	GEMINI_202602: {
		ID:          GEMINI_202602,
		Title:       "Gemini Structured Output Profile",
		Description: "Meta-schema for JSON Schemas intended for Gemini structured output",
		Baseline:    "202602",
		SchemaFile:  "schemas/gemini_202602.yaml",
	},
	GEMINI_202503: {
		ID:          GEMINI_202503,
		Title:       "Gemini 2.0 Structured Output Profile",
		Description: "Meta-schema for JSON Schemas intended for Gemini 2.0 structured output with required propertyOrdering",
		Baseline:    "202602",
		SchemaFile:  "schemas/gemini_202503.yaml",
	},
	MINIMAL_202602: {
		ID:          MINIMAL_202602,
		Title:       "Minimal Gemini/OpenAI Common Profile",
		Description: "Lowest common denominator profile across the documented Gemini and OpenAI structured-output JSON Schema subsets",
		Baseline:    "202602",
		SchemaFile:  "schemas/minimal_202602.yaml",
	},
}

// ListProfiles returns all registered profile infos.
func ListProfiles() []ProfileInfo {
	out := make([]ProfileInfo, 0, len(profiles))
	// Stable order
	for _, id := range []ProfileID{OPENAI_202602, GEMINI_202602, GEMINI_202503, MINIMAL_202602} {
		out = append(out, profiles[id])
	}
	return out
}

// GetProfileInfo returns the info for a profile ID.
func GetProfileInfo(id ProfileID) (ProfileInfo, error) {
	p, ok := profiles[id]
	if !ok {
		return ProfileInfo{}, fmt.Errorf("unknown profile: %s", id)
	}
	return p, nil
}

// GetProfileSchema returns the raw YAML bytes of the embedded meta-schema for a profile.
func GetProfileSchema(id ProfileID) ([]byte, error) {
	p, ok := profiles[id]
	if !ok {
		return nil, fmt.Errorf("unknown profile: %s", id)
	}
	return schemaFS.ReadFile(p.SchemaFile)
}
