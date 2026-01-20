package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/heywinit/grechen/internal/cli"
	"github.com/heywinit/grechen/internal/extract"
	"github.com/heywinit/grechen/internal/llm"
	"github.com/heywinit/grechen/internal/patterns"
	"github.com/heywinit/grechen/internal/rules"
	"github.com/heywinit/grechen/internal/stats"
	"github.com/heywinit/grechen/internal/store"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists (ignore errors if file doesn't exist)
	_ = godotenv.Load()

	// Get data directory (default: ~/.grechen)
	dataDir := os.Getenv("GRECHEN_DATA_DIR")
	if dataDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: failed to get home directory: %v\n", err)
			os.Exit(1)
		}
		dataDir = filepath.Join(home, ".grechen")
	}

	// Initialize store
	s, err := store.New(dataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to initialize store: %v\n", err)
		os.Exit(1)
	}

	// Initialize LLM provider (default: gemini)
	llmProvider := getLLMProvider()
	if llmProvider == nil {
		fmt.Fprintf(os.Stderr, "error: failed to initialize LLM provider\n")
		os.Exit(1)
	}
	extractor := extract.NewLLMExtractor(llmProvider)

	// Initialize components
	r := rules.New(s)
	st := stats.New(s)
	p := patterns.New(s, st)
	c := cli.New(s, extractor, r, st, p)

	// Parse arguments
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "usage: grechen <command> [args...]\n")
		fmt.Fprintf(os.Stderr, "commands: setup | <natural language> | goodnight | review | today | commitments | thats-wrong\n")
		os.Exit(1)
	}

	command := args[0]

	// Route to appropriate handler
	var handlerErr error
	switch command {
	case "setup":
		handlerErr = c.HandleSetup()
	case "goodnight":
		handlerErr = c.HandleGoodnight()
	case "review":
		handlerErr = c.HandleReview()
	case "today":
		handlerErr = c.HandleToday()
	case "commitments":
		handlerErr = c.HandleCommitments()
	case "thats-wrong":
		handlerErr = c.HandleThatsWrong()
	default:
		// Treat as natural language input
		input := strings.Join(args, " ")
		handlerErr = c.HandleInput(input)
	}

	if handlerErr != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", handlerErr)
		os.Exit(1)
	}
}

func getLLMProvider() llm.Provider {
	providerType := os.Getenv("GRECHEN_LLM_PROVIDER")
	if providerType == "" {
		providerType = "gemini"
	}

	switch providerType {
	case "gemini":
		provider, err := llm.NewGeminiProvider()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: failed to initialize gemini provider: %v\n", err)
			fmt.Fprintf(os.Stderr, "hint: set GEMINI_API_KEY environment variable\n")
			return nil
		}
		return provider
	default:
		fmt.Fprintf(os.Stderr, "error: unknown provider %s (only 'gemini' is supported)\n", providerType)
		return nil
	}
}
