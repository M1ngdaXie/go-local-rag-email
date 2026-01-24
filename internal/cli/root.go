package cli

import (
	"github.com/M1ngdaXie/go-local-rag-email/internal/app"
	"github.com/spf13/cobra"
)

var (
	application *app.App
	cfgFile     string
	verbose     bool
)

// Execute runs the CLI
func Execute(app *app.App) error {
	application = app
	return rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use:   "go-local-rag-email",
	Short: "Local-first RAG email assistant",
	Long: `A CLI/TUI application for managing emails with RAG-based search and AI summarization.

Features:
  - Sync emails from Gmail
  - Semantic search with natural language
  - AI-powered email summaries
  - Interactive TUI interface`,
	Version: "0.1.0",
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.go-local-rag-email/config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.AddCommand(testEmailCmd)
}
