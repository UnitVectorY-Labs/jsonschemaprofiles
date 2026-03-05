package jsonschemaprofiles

// validatePhase2 dispatches to profile-specific phase 2 checks.
func validatePhase2(profileID ProfileID, candidateMap map[string]interface{}, rawBytes []byte, report *Report, opts *ValidateOptions) {
	if candidateMap == nil {
		return
	}

	switch profileID {
	case OPENAI_202602:
		validatePhase2OpenAI(candidateMap, rawBytes, report, opts)
	case GEMINI_202602:
		validatePhase2Gemini(candidateMap, report, false)
	case GEMINI_2_0_202602:
		validatePhase2Gemini(candidateMap, report, true)
	case MINIMAL_202602:
		validatePhase2Minimal(candidateMap, rawBytes, report)
	}
}
