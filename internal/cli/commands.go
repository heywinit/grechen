package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/heywinit/grechen/internal/core"
)

// HandleToday shows situational awareness for today
func (c *CLI) HandleToday() error {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// Get today's stats
	stats, err := c.stats.ComputeDailyStats(today)
	if err != nil {
		return err
	}

	fmt.Printf("today (%s)\n", today.Format("2006-01-02"))
	fmt.Printf("  logs: %d\n", stats.LogCount)
	if stats.WorkStartTime != nil {
		fmt.Printf("  work started: %s\n", stats.WorkStartTime.Format("15:04"))
	}
	fmt.Printf("  progress entries: %d\n", stats.ProgressEntries)
	fmt.Printf("  commitment updates: %d\n", stats.CommitmentUpdates)

	// Show today's logs
	content, err := c.store.ReadDailyFile(today)
	if err != nil {
		return err
	}
	logs := extractLogs(content)
	if len(logs) > 0 {
		fmt.Println("\nlogs:")
		for _, log := range logs {
			fmt.Printf("  %s\n", log)
		}
	}

	// Show open commitments
	commitments, err := c.store.ListOpenCommitments()
	if err != nil {
		return err
	}

	if len(commitments) > 0 {
		fmt.Println("\nopen commitments:")
		for _, c := range commitments {
			daysUntil := int(c.Expectation.Deadline.Sub(now).Hours() / 24)
			fmt.Printf("  %s → %s (due %s, %d days)\n",
				c.PersonID,
				c.Expectation.Description,
				c.Expectation.Deadline.Format("2006-01-02"),
				daysUntil)
		}
	}

	return nil
}

// extractLogs extracts log entries from daily file content
func extractLogs(content string) []string {
	lines := strings.Split(content, "\n")
	inLogsSection := false
	var logs []string

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
			// Format: "- HHMM text" -> display as "HH:MM text"
			if len(trimmed) > 2 {
				logText := trimmed[2:] // Remove "- "
				// Try to format time if it starts with 4 digits (HHMM format)
				if len(logText) >= 4 {
					timePart := logText[:4]
					rest := logText[4:]
					// Check if first 4 chars are digits
					if len(timePart) == 4 && strings.Trim(timePart, "0123456789") == "" {
						// Format as HH:MM
						formatted := timePart[:2] + ":" + timePart[2:]
						logs = append(logs, formatted+" "+strings.TrimSpace(rest))
						continue
					}
				}
				// If no time format detected, just show the log as-is
				logs = append(logs, strings.TrimSpace(logText))
			}
		}
	}

	return logs
}

// HandleCommitments shows all commitments
func (c *CLI) HandleCommitments() error {
	commitments, err := c.store.ListCommitments()
	if err != nil {
		return err
	}

	if len(commitments) == 0 {
		fmt.Println("no commitments")
		return nil
	}

	fmt.Println("commitments:")
	for _, c := range commitments {
		fmt.Printf("  [%s] %s → %s (due %s, status: %s)\n",
			c.ID,
			c.PersonID,
			c.Expectation.Description,
			c.Expectation.Deadline.Format("2006-01-02"),
			c.Status)
	}

	return nil
}

// HandleReview shows stats summary and pattern alerts
func (c *CLI) HandleReview() error {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// Get rolling stats (last 7 days)
	rollingStats, err := c.stats.ComputeRollingStats(today, 7)
	if err != nil {
		return err
	}

	fmt.Println("last 7 days:")
	fmt.Printf("  avg logs/day: %.1f\n", rollingStats.AvgLogCount)
	if rollingStats.AvgWorkStartTime != nil {
		fmt.Printf("  avg work start: %s\n", rollingStats.AvgWorkStartTime.Format("15:04"))
	}
	fmt.Printf("  avg progress entries/day: %.1f\n", rollingStats.AvgProgressEntries)
	fmt.Printf("  avg commitment updates/day: %.1f\n", rollingStats.AvgCommitmentUpdates)

	// Check patterns
	deviations, err := c.patterns.Evaluate(today, rollingStats)
	if err != nil {
		return err
	}

	if len(deviations) > 0 {
		fmt.Println("\npatterns:")
		for _, dev := range deviations {
			fmt.Printf("  [%s] %s: %s\n", dev.Severity, dev.Pattern, dev.Question.Text)
		}
	}

	return nil
}

// HandleThatsWrong handles correction flow
func (c *CLI) HandleThatsWrong() error {
	fmt.Println("what was wrong?")
	// In future: read from stdin or use Charm form
	// For now, this is a placeholder
	return fmt.Errorf("correction flow not fully implemented")
}

// HandleTodo shows all remaining todos from previous days
func (c *CLI) HandleTodo() error {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	commitments, err := c.store.ListOpenCommitmentsFromPreviousDays(today)
	if err != nil {
		return err
	}

	if len(commitments) == 0 {
		fmt.Println("no remaining todos from previous days")
		return nil
	}

	fmt.Printf("remaining todos (%d):\n", len(commitments))
	for i, c := range commitments {
		daysAgo := int(today.Sub(c.CreatedAt).Hours() / 24)
		daysUntil := int(c.Expectation.Deadline.Sub(now).Hours() / 24)
		fmt.Printf("  %d. %s → %s (created %d days ago, due %s, %d days left)\n",
			i+1,
			c.PersonID,
			c.Expectation.Description,
			daysAgo,
			c.Expectation.Deadline.Format("2006-01-02"),
			daysUntil)
		if c.ProjectID != "" {
			fmt.Printf("     project: %s\n", c.ProjectID)
		}
	}

	return nil
}

// HandleProjects shows projects and allows editing
func (c *CLI) HandleProjects() error {
	projects, err := c.store.ListProjects()
	if err != nil {
		return err
	}

	if len(projects) == 0 {
		fmt.Println("no projects")
		fmt.Println("projects are created automatically when mentioned in commitments")
		return nil
	}

	fmt.Println("projects:")
	for i, p := range projects {
		priority := "normal"
		if p.Priority > 0 {
			priority = "high"
		} else if p.Priority < 0 {
			priority = "low"
		}
		fmt.Printf("  %d. %s (priority: %s)\n", i+1, p.ID, priority)
		if len(p.Metadata) > 0 {
			fmt.Printf("     metadata: %v\n", p.Metadata)
		}
	}

	// Check if user wants to edit
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\nedit project? (enter number or 'n' to skip): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "n" || input == "" {
		return nil
	}

	idx, err := strconv.Atoi(input)
	if err != nil || idx < 1 || idx > len(projects) {
		fmt.Println("invalid selection")
		return nil
	}

	project := projects[idx-1]
	return c.editProject(project, reader)
}

func (c *CLI) editProject(project *core.Project, reader *bufio.Reader) error {
	fmt.Printf("\nediting project: %s\n", project.ID)
	fmt.Println("(press enter to keep current value)")

	// Edit priority
	fmt.Printf("priority (current: %d, enter number or 'high'/'normal'/'low'): ", project.Priority)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input != "" {
		switch input {
		case "high":
			project.Priority = 1
		case "low":
			project.Priority = -1
		case "normal", "0":
			project.Priority = 0
		default:
			if prio, err := strconv.Atoi(input); err == nil {
				project.Priority = prio
			}
		}
	}

	// Save
	if err := c.store.SaveProject(project); err != nil {
		return fmt.Errorf("failed to save project: %w", err)
	}

	fmt.Printf("updated project: %s (priority: %d)\n", project.ID, project.Priority)
	return nil
}

// HandlePeople shows people and allows editing
func (c *CLI) HandlePeople() error {
	people, err := c.store.ListPeople()
	if err != nil {
		return err
	}

	if len(people) == 0 {
		fmt.Println("no people")
		fmt.Println("people are created automatically when mentioned in commitments")
		return nil
	}

	fmt.Println("people:")
	for i, p := range people {
		fmt.Printf("  %d. %s (id: %s)\n", i+1, p.Name, p.ID)
		if len(p.Metadata) > 0 {
			fmt.Printf("     metadata: %v\n", p.Metadata)
		}
	}

	// Check if user wants to edit
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\nedit person? (enter number or 'n' to skip): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "n" || input == "" {
		return nil
	}

	idx, err := strconv.Atoi(input)
	if err != nil || idx < 1 || idx > len(people) {
		fmt.Println("invalid selection")
		return nil
	}

	person := people[idx-1]
	return c.editPerson(person, reader)
}

func (c *CLI) editPerson(person *core.Person, reader *bufio.Reader) error {
	fmt.Printf("\nediting person: %s\n", person.ID)
	fmt.Println("(press enter to keep current value)")

	// Edit name
	fmt.Printf("name (current: %s): ", person.Name)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input != "" {
		person.Name = input
	}

	// Initialize metadata if needed
	if person.Metadata == nil {
		person.Metadata = make(map[string]any)
	}

	// Save
	if err := c.store.SavePerson(person); err != nil {
		return fmt.Errorf("failed to save person: %w", err)
	}

	fmt.Printf("updated person: %s (name: %s)\n", person.ID, person.Name)
	return nil
}
