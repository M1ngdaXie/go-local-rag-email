package config

import (
	"fmt"
	"os"
)

// Validate checks if the configuration is valid
func Validate(cfg *Config) error {
	// Check OpenAI API key exists
	if cfg.OpenAI.APIKey == "" {
		return fmt.Errorf("openai.api_key is required (set OPENAI_API_KEY environment variable)")
	}

	// Create data directory if it doesn't exist
	if err := os.MkdirAll(cfg.App.DataDir, 0755); err != nil {
		return fmt.Errorf("cannot create data directory: %w", err)
	}

	// Set defaults for optional fields
	if cfg.OpenAI.EmbeddingModel == "" {
		cfg.OpenAI.EmbeddingModel = "text-embedding-3-small"
	}
	if cfg.OpenAI.ChatModel == "" {
		cfg.OpenAI.ChatModel = "gpt-4o-mini"
	}

	// Validate Qdrant settings
	if cfg.Qdrant.URL == "" {
		return fmt.Errorf("qdrant.url is required")
	}
	if cfg.Qdrant.VectorSize != 1536 && cfg.Qdrant.VectorSize != 3072 {
		return fmt.Errorf("qdrant.vector_size must be 1536 or 3072 (got %d)", cfg.Qdrant.VectorSize)
	}

	return nil
}
