package stats

import (
	"fmt"
	"strings"
	"time"
)

func countLogEntries(content string) int {
	lines := strings.Split(content, "\n")
	inLogsSection := false
	count := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "## logs" {
			inLogsSection = true
			continue
		}
		if strings.HasPrefix(trimmed, "## ") {
			inLogsSection = false
			continue
		}
		if inLogsSection && strings.HasPrefix(trimmed, "- ") {
			count++
		}
	}

	return count
}

func findWorkStartTime(content string, date time.Time) *time.Time {
	lines := strings.Split(content, "\n")
	inLogsSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "## logs" {
			inLogsSection = true
			continue
		}
		if strings.HasPrefix(trimmed, "## ") {
			inLogsSection = false
			continue
		}
		if inLogsSection && strings.HasPrefix(trimmed, "- ") {
			// Check if this looks like a work/progress entry
			lower := strings.ToLower(trimmed)
			if strings.Contains(lower, "work") || strings.Contains(lower, "start") || 
			   strings.Contains(lower, "coding") || strings.Contains(lower, "sitting") {
				// Try to extract time from entry
				// Format: "- HHMM text"
				if len(trimmed) > 6 && trimmed[2:6] != "" {
					timeStr := trimmed[2:6]
					if len(timeStr) == 4 {
						hour := 0
						min := 0
						fmt.Sscanf(timeStr, "%2d%2d", &hour, &min)
						if hour >= 0 && hour < 24 && min >= 0 && min < 60 {
							t := time.Date(date.Year(), date.Month(), date.Day(), hour, min, 0, 0, time.UTC)
							return &t
						}
					}
				}
			}
		}
	}

	return nil
}

func countProgressEntries(content string) int {
	lines := strings.Split(content, "\n")
	count := 0

	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "done") || strings.Contains(lower, "finished") || 
		   strings.Contains(lower, "completed") || strings.Contains(lower, "progress") {
			count++
		}
	}

	return count
}

func countCommitmentUpdates(content string) int {
	lines := strings.Split(content, "\n")
	inCommitmentsSection := false
	count := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "## commitments" {
			inCommitmentsSection = true
			continue
		}
		if strings.HasPrefix(trimmed, "## ") {
			inCommitmentsSection = false
			continue
		}
		if inCommitmentsSection && strings.HasPrefix(trimmed, "- ") {
			count++
		}
	}

	return count
}
