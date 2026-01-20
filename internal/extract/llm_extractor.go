package extract

import (
	"fmt"

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
	// Get JSON from LLM
	jsonData, err := e.provider.ExtractJSON(input)
	if err != nil {
		return nil, nil, fmt.Errorf("llm extraction failed: %w", err)
	}

	// Validate and parse
	candidate, err := ValidateCandidate(jsonData)
	if err != nil {
		return nil, nil, fmt.Errorf("validation failed: %w", err)
	}

	// Return candidate and any questions
	return []core.Candidate{*candidate}, candidate.Questions, nil
}
