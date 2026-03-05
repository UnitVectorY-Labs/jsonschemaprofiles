package engine

// validatePhase2 dispatches to profile-specific phase 2 checks.
func validatePhase2(profileID string, candidateMap map[string]interface{}, rawBytes []byte, report *Report, opts *ValidateOptions) {
	if candidateMap == nil {
		return
	}

	switch profileID {
	case profileOpenAI:
		validatePhase2OpenAI(candidateMap, rawBytes, report, opts)
	case profileGemini:
		validatePhase2Gemini(candidateMap, report, false)
	case profileGemini20:
		validatePhase2Gemini(candidateMap, report, true)
	case profileMinimal:
		validatePhase2Minimal(candidateMap, rawBytes, report)
	}
}
