package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

// Load reads and validates the configuration
func Load() (*Config, error) {
	v := viper.New()

	// Set config file name and type
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Add config search paths
	v.AddConfigPath("./configs")
	v.AddConfigPath("$HOME/.go-local-rag-email")
	v.AddConfigPath(".")

	// Set defaults
	setDefaults(v)

	// Enable environment variable overrides
	v.AutomaticEnv()
	v.SetEnvPrefix("RAGMAIL")

	// Read config file (it's okay if it doesn't exist, we'll use defaults)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
		// Config file not found, use defaults
	}

	// Unmarshal into our Config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Expand ~ in paths
	if err := expandPaths(&cfg); err != nil {
		return nil, fmt.Errorf("failed to expand paths: %w", err)
	}

	// Validate the configuration
	if err := Validate(&cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default values for all config options
func setDefaults(v *viper.Viper) {
	// App defaults
	v.SetDefault("app.name", "go-local-rag-email")
	v.SetDefault("app.data_dir", "~/.go-local-rag-email")

	// Gmail defaults
	v.SetDefault("gmail.credentials_path", "configs/credentials.json")
	v.SetDefault("gmail.token_path", "~/.go-local-rag-email/token.json")
	v.SetDefault("gmail.scopes", []string{"https://www.googleapis.com/auth/gmail.readonly"})

	// OpenAI defaults
	v.SetDefault("openai.embedding_model", "text-embedding-3-small")
	v.SetDefault("openai.chat_model", "gpt-4o-mini")
	v.SetDefault("openai.max_tokens", 2000)
	v.SetDefault("openai.temperature", 0.7)

	// SQLite defaults
	v.SetDefault("sqlite.path", "~/.go-local-rag-email/emails.db")
	v.SetDefault("sqlite.max_open_conns", 10)
	v.SetDefault("sqlite.max_idle_conns", 5)
	v.SetDefault("sqlite.conn_max_lifetime", "1h")
	v.SetDefault("sqlite.enable_wal", true)
	v.SetDefault("sqlite.enable_foreign_keys", true)

	// Qdrant defaults
	v.SetDefault("qdrant.url", "http://localhost:6333")
	v.SetDefault("qdrant.collection_name", "email_embeddings")
	v.SetDefault("qdrant.vector_size", 1536)
	v.SetDefault("qdrant.distance", "Cosine")

	// Logging defaults
	v.SetDefault("logging.level", "info")
}

// expandPaths expands ~ to home directory in all path fields
func expandPaths(cfg *Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Helper function to expand a single path
	expand := func(path string) string {
		if len(path) > 0 && path[0] == '~' {
			return filepath.Join(homeDir, path[1:])
		}
		return path
	}

	// Expand all path fields
	cfg.App.DataDir = expand(cfg.App.DataDir)
	cfg.Gmail.CredentialsPath = expand(cfg.Gmail.CredentialsPath)
	cfg.Gmail.TokenPath = expand(cfg.Gmail.TokenPath)
	cfg.SQLite.Path = expand(cfg.SQLite.Path)
	cfg.Logging.FilePath = expand(cfg.Logging.FilePath)

	// Parse duration string for SQLite
	if cfg.SQLite.ConnMaxLifetime == 0 {
		cfg.SQLite.ConnMaxLifetime = time.Hour
	}

	return nil
}
