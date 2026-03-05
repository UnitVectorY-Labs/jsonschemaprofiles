package jsonschemaprofiles

import (
	"testing"
)

func TestListProfiles(t *testing.T) {
	profiles := ListProfiles()
	if len(profiles) != 4 {
		t.Errorf("expected 4 profiles, got %d", len(profiles))
	}

	// Verify order
	expectedIDs := []ProfileID{OPENAI_202602, GEMINI_202602, GEMINI_2_0_202602, MINIMAL_202602}
	for i, p := range profiles {
		if p.ID != expectedIDs[i] {
			t.Errorf("expected profile %d to be %s, got %s", i, expectedIDs[i], p.ID)
		}
	}
}

func TestGetProfileInfo(t *testing.T) {
	info, err := GetProfileInfo(OPENAI_202602)
	if err != nil {
		t.Fatalf("GetProfileInfo error: %v", err)
	}
	if info.ID != OPENAI_202602 {
		t.Errorf("expected ID %s, got %s", OPENAI_202602, info.ID)
	}
	if info.Baseline != "202602" {
		t.Errorf("expected baseline 202602, got %s", info.Baseline)
	}
}

func TestGetProfileInfoUnknown(t *testing.T) {
	_, err := GetProfileInfo("UNKNOWN_PROFILE")
	if err == nil {
		t.Error("expected error for unknown profile")
	}
}

func TestGetProfileSchema(t *testing.T) {
	for _, id := range []ProfileID{OPENAI_202602, GEMINI_202602, GEMINI_2_0_202602, MINIMAL_202602} {
		data, err := GetProfileSchema(id)
		if err != nil {
			t.Errorf("GetProfileSchema(%s) error: %v", id, err)
			continue
		}
		if len(data) == 0 {
			t.Errorf("GetProfileSchema(%s) returned empty data", id)
		}
	}
}

func TestGetProfileSchemaUnknown(t *testing.T) {
	_, err := GetProfileSchema("UNKNOWN_PROFILE")
	if err == nil {
		t.Error("expected error for unknown profile")
	}
}
