package extract

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/heywinit/grechen/internal/core"
)

// ValidateCandidate validates and filters candidates based on confidence and schema
func ValidateCandidate(rawJSON []byte) (*core.Candidate, error) {
	var candidate core.Candidate
	if err := json.Unmarshal(rawJSON, &candidate); err != nil {
		jsonStr := string(rawJSON)
		// Show more context for debugging
		displayStr := jsonStr
		if len(displayStr) > 1000 {
			displayStr = displayStr[:1000] + "... (truncated for display)"
		}
		return nil, fmt.Errorf("invalid JSON: %w\n  received: %q", err, displayStr)
	}

	// Check confidence threshold
	if candidate.Confidence < MinConfidence {
		return nil, fmt.Errorf("confidence too low: %.2f < %.2f", candidate.Confidence, MinConfidence)
	}

	// Validate intent type
	validTypes := map[core.IntentType]bool{
		core.IntentLog:        true,
		core.IntentProgress:   true,
		core.IntentCommitment: true,
		core.IntentUpdate:     true,
		core.IntentEvent:      true,
		core.IntentCorrection: true,
	}
	if !validTypes[candidate.Type] {
		return nil, fmt.Errorf("invalid intent type: %s", candidate.Type)
	}

	// Validate data structure based on type
	if err := validateCandidateData(candidate.Type, candidate.Data); err != nil {
		return nil, fmt.Errorf("invalid data structure: %w", err)
	}

	return &candidate, nil
}

func validateCandidateData(intentType core.IntentType, data map[string]any) error {
	switch intentType {
	case core.IntentCommitment:
		return validateCommitmentData(data)
	case core.IntentUpdate:
		return validateUpdateData(data)
	case core.IntentEvent:
		return validateEventData(data)
	case core.IntentProgress:
		return validateProgressData(data)
	case core.IntentLog, core.IntentCorrection:
		// These types have minimal structure requirements
		return nil
	default:
		return fmt.Errorf("unknown intent type: %s", intentType)
	}
}

func validateCommitmentData(data map[string]any) error {
	// Check required fields
	if _, ok := data["person"]; !ok {
		return fmt.Errorf("missing required field: person")
	}
	if _, ok := data["expectation"]; !ok {
		return fmt.Errorf("missing required field: expectation")
	}

	// Validate expectation structure
	expRaw, ok := data["expectation"].(map[string]any)
	if !ok {
		return fmt.Errorf("expectation must be an object")
	}

	if _, ok := expRaw["description"]; !ok {
		return fmt.Errorf("expectation missing description")
	}
	if _, ok := expRaw["deadline"]; !ok {
		return fmt.Errorf("expectation missing deadline")
	}

	// Validate deadline format
	deadlineStr, ok := expRaw["deadline"].(string)
	if !ok {
		return fmt.Errorf("deadline must be a string")
	}
	if _, err := time.Parse("2006-01-02", deadlineStr); err != nil {
		return fmt.Errorf("invalid deadline format: %w", err)
	}

	// Validate hardness if present
	if hardness, ok := expRaw["hardness"].(string); ok {
		if hardness != "hard" && hardness != "soft" {
			return fmt.Errorf("hardness must be 'hard' or 'soft'")
		}
	}

	return nil
}

func validateUpdateData(data map[string]any) error {
	if _, ok := data["commitment_id"]; !ok {
		// If no commitment_id, must have person or project to identify
		if _, hasPerson := data["person"]; !hasPerson {
			if _, hasProject := data["project"]; !hasProject {
				return fmt.Errorf("update must have commitment_id or person/project")
			}
		}
	}
	return nil
}

func validateEventData(data map[string]any) error {
	if _, ok := data["time"]; !ok {
		return fmt.Errorf("event missing time")
	}
	return nil
}

func validateProgressData(data map[string]any) error {
	if _, ok := data["project"]; !ok {
		return fmt.Errorf("progress missing project")
	}
	return nil
}
