package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/heywinit/grechen/internal/core"
)

func (s *Store) AppendLog(date time.Time, entry *core.Entry) error {
	filename := s.dailyFilename(date)
	return s.appendToSection(filename, "logs", formatLogEntry(entry))
}

func (s *Store) AppendCommitment(date time.Time, commitment *core.Commitment) error {
	filename := s.dailyFilename(date)
	line := formatCommitment(commitment)
	return s.appendToSection(filename, "commitments", line)
}

func (s *Store) AppendNote(date time.Time, text string) error {
	filename := s.dailyFilename(date)
	return s.appendToSection(filename, "notes", text)
}

func (s *Store) ReadDailyFile(date time.Time) (string, error) {
	filename := s.dailyFilename(date)
	data, err := os.ReadFile(filename)
	if os.IsNotExist(err) {
		return "", nil
	}
	return string(data), err
}

func (s *Store) dailyFilename(date time.Time) string {
	return filepath.Join(s.DailyDir(), date.Format("2006-01-02")+".md")
}

func (s *Store) appendToSection(filename, section, content string) error {
	// Read existing file
	existing, err := os.ReadFile(filename)
	fileExists := err == nil

	var lines []string
	if fileExists {
		lines = strings.Split(string(existing), "\n")
	}

	// Find or create section
	sectionHeader := fmt.Sprintf("## %s", section)
	sectionIndex := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == sectionHeader {
			sectionIndex = i
			break
		}
	}

	// If section doesn't exist, create it
	if sectionIndex == -1 {
		if len(lines) > 0 && lines[len(lines)-1] != "" {
			lines = append(lines, "")
		}
		lines = append(lines, sectionHeader)
		lines = append(lines, "")
		sectionIndex = len(lines) - 1
	}

	// Append content after section header
	insertIndex := sectionIndex + 1
	// Skip empty line after header if it exists
	if insertIndex < len(lines) && strings.TrimSpace(lines[insertIndex]) == "" {
		insertIndex++
	}

	// Insert content
	newLines := make([]string, 0, len(lines)+1)
	newLines = append(newLines, lines[:insertIndex]...)
	newLines = append(newLines, content)
	newLines = append(newLines, lines[insertIndex:]...)

	// Write file
	return os.WriteFile(filename, []byte(strings.Join(newLines, "\n")), 0644)
}

func formatLogEntry(entry *core.Entry) string {
	return fmt.Sprintf("- %s %s", entry.Timestamp.Format("1504"), entry.Raw)
}

func formatCommitment(commitment *core.Commitment) string {
	deadline := commitment.Expectation.Deadline.Format("2006-01-02")
	return fmt.Sprintf("- %s → %s (due %s)", commitment.PersonID, commitment.Expectation.Description, deadline)
}
