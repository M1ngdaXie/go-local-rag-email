package config

import (
	"fmt"
	"os"
)

// Validate checks if the configuration is valid.
// It does NOT set defaults. Missing required fields will cause errors.
func Validate(cfg *Config) error {
	// ---- OpenAI ----
	if cfg.OpenAI.APIKey == "" {
		return fmt.Errorf("openai.api_key is required (set OPENAI_API_KEY environment variable)")
	}

	if cfg.OpenAI.EmbeddingModel == "" {
		return fmt.Errorf("openai.embedding_model is required")
	}

	if cfg.OpenAI.ChatModel == "" {
		return fmt.Errorf("openai.chat_model is required")
	}

	// ---- App ----
	if cfg.App.DataDir == "" {
		return fmt.Errorf("app.data_dir is required")
	}

	if err := os.MkdirAll(cfg.App.DataDir, 0755); err != nil {
		return fmt.Errorf("cannot create data directory %q: %w", cfg.App.DataDir, err)
	}

	// ---- Qdrant ----
	if cfg.Qdrant.URL == "" {
		return fmt.Errorf("qdrant.url is required")
	}

	if cfg.Qdrant.VectorSize == 0 {
		return fmt.Errorf("qdrant.vector_size is required")
	}

	if cfg.Qdrant.VectorSize != 1536 && cfg.Qdrant.VectorSize != 3072 {
		return fmt.Errorf(
			"qdrant.vector_size must be 1536 or 3072 (got %d)",
			cfg.Qdrant.VectorSize,
		)
	}

	return nil
}
