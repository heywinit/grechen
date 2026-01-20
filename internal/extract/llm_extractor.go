package extract

import (
	"fmt"
	"time"

	"github.com/heywinit/grechen/internal/core"
	"github.com/heywinit/grechen/internal/llm"
)

// LLMExtractor uses an LLM provider to extract candidates
type LLMExtractor struct {
	provider llm.Provider
}

func NewLLMExtractor(p llm.Provider) *LLMExtractor {
	return &LLMExtractor{provider: p}
}

func (e *LLMExtractor) Extract(input string) ([]core.Candidate, []core.Question, error) {
	// Get JSON from LLM (pass current time for date calculations)
	jsonData, err := e.provider.ExtractJSON(input, time.Now())
	if err != nil {
		return nil, nil, fmt.Errorf("llm extraction failed: %w", err)
	}

	// Validate and parse
	candidate, err := ValidateCandidate(jsonData)
	if err != nil {
		// Include raw JSON in error for debugging
		jsonStr := string(jsonData)
		jsonPreview := jsonStr
		if len(jsonPreview) > 500 {
			jsonPreview = jsonPreview[:500] + "... (truncated)"
		}
		return nil, nil, fmt.Errorf("validation failed: %w\n  input: %q\n  extracted JSON (%d bytes): %s", err, input, len(jsonData), jsonPreview)
	}

	// Return candidate and any questions
	return []core.Candidate{*candidate}, candidate.Questions, nil
}
