package llm

import (
	"encoding/json"
	"time"
)

// MockProvider is a simple mock implementation for testing
type MockProvider struct {
	responses map[string][]byte
}

func NewMockProvider() *MockProvider {
	return &MockProvider{
		responses: make(map[string][]byte),
	}
}

func (m *MockProvider) ExtractJSON(input string) ([]byte, error) {
	// Check for predefined response
	if resp, ok := m.responses[input]; ok {
		return resp, nil
	}

	// Default mock response - simple commitment extraction
	if contains(input, "told") || contains(input, "promised") || contains(input, "will") {
		response := map[string]any{
			"type":       "commitment",
			"confidence": 0.85,
			"data": map[string]any{
				"person": "prof",
				"expectation": map[string]any{
					"description": "pr ready",
					"deadline":     time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
					"hardness":     "hard",
				},
			},
			"questions": []any{},
		}
		return json.Marshal(response)
	}

	// Default: log entry
	response := map[string]any{
		"type":       "log",
		"confidence": 0.90,
		"data": map[string]any{
			"text": input,
		},
		"questions": []any{},
	}
	return json.Marshal(response)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
