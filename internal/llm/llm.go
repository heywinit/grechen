package llm

import "time"

// Provider is an abstraction for LLM providers
type Provider interface {
	// ExtractJSON takes natural language input and returns structured JSON
	// The JSON should match the Candidate schema from the extract package
	// now is the current time, used for relative date calculations
	ExtractJSON(input string, now time.Time) ([]byte, error)
}
