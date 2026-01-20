package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

// HandleSetup initializes the basic directory structure and creates initial files
func (c *CLI) HandleSetup() error {
	fmt.Println("setting up grechen...")

	// Ensure directories exist (store.New() in main already does this, but ensure here too)
	dailyDir := c.store.DailyDir()
	metaDir := c.store.MetaDir()

	if err := os.MkdirAll(dailyDir, 0755); err != nil {
		return fmt.Errorf("failed to create daily directory: %w", err)
	}
	if err := os.MkdirAll(metaDir, 0755); err != nil {
		return fmt.Errorf("failed to create meta directory: %w", err)
	}
	fmt.Println("  created directory structure")

	// Create empty JSON files if they don't exist
	files := map[string][]byte{
		"people.json":      []byte("[]\n"),
		"projects.json":    []byte("[]\n"),
		"commitments.json": []byte("[]\n"),
	}

	for filename, content := range files {
		filepath := filepath.Join(metaDir, filename)
		if _, err := os.Stat(filepath); os.IsNotExist(err) {
			if err := os.WriteFile(filepath, content, 0644); err != nil {
				return fmt.Errorf("failed to create %s: %w", filename, err)
			}
			fmt.Printf("  created %s\n", filename)
		} else {
			fmt.Printf("  %s already exists, skipping\n", filename)
		}
	}

	// Create a README in the data directory to explain the structure
	readmePath := filepath.Join(c.store.DataDir(), "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		readmeContent := "# Grechen Data Directory\n\n" +
			"## Structure\n\n" +
			"- `daily/` - Daily markdown files (YYYY-MM-DD.md)\n" +
			"  - Each file contains sections: ## logs, ## commitments, ## notes\n" +
			"  - Files are append-only, never rewritten\n\n" +
			"- `meta/` - Metadata storage (JSON files)\n" +
			"  - `people.json` - People you interact with\n" +
			"  - `projects.json` - Projects you work on\n" +
			"  - `commitments.json` - All commitments and their history\n\n" +
			"## Adding Data\n\n" +
			"People and projects are automatically created when you mention them in your logs:\n\n" +
			"```\n" +
			"grechen met with alice to discuss the project\n" +
			"grechen worked on project-x for 2 hours\n" +
			"```\n\n" +
			"You can also manually edit the JSON files in `meta/` if needed.\n"
		if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
			return fmt.Errorf("failed to create README: %w", err)
		}
		fmt.Println("  created README.md")
	}

	fmt.Println("\nsetup complete!")
	fmt.Printf("data directory: %s\n", c.store.DataDir())
	fmt.Println("\nnext steps:")
	fmt.Println("  - add people by mentioning them in your logs:")
	fmt.Println("    grechen met with alice to discuss the project")
	fmt.Println("  - add projects by mentioning them in your logs:")
	fmt.Println("    grechen worked on project-x for 2 hours")
	fmt.Println("  - start logging your activities:")
	fmt.Println("    grechen had lunch. gonna sit for working")
	fmt.Println("  - create commitments:")
	fmt.Println("    grechen told bob the report will be ready by friday")
	fmt.Println("\ncommands:")
	fmt.Println("  grechen today          - see today's summary")
	fmt.Println("  grechen commitments    - list all commitments")
	fmt.Println("  grechen goodnight      - end-of-day review")

	return nil
}
