package llm

// Provider is an abstraction for LLM providers
type Provider interface {
	// ExtractJSON takes natural language input and returns structured JSON
	// The JSON should match the Candidate schema from the extract package
	ExtractJSON(input string) ([]byte, error)
}
