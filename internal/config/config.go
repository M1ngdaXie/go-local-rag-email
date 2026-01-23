package config

import "time"

// Config holds all application configuration
type Config struct {
	App     AppConfig
	Gmail   GmailConfig
	OpenAI  OpenAIConfig
	SQLite  SQLiteConfig
	Qdrant  QdrantConfig
	Logging LoggingConfig
}

// AppConfig holds application-level settings
type AppConfig struct {
	Name    string
	DataDir string `mapstructure:"data_dir"`
}

// GmailConfig holds Gmail API settings
type GmailConfig struct {
	CredentialsPath string   `mapstructure:"credentials_path"`
	TokenPath       string   `mapstructure:"token_path"`
	Scopes          []string `mapstructure:"scopes"`
}

// OpenAIConfig holds OpenAI API settings
type OpenAIConfig struct {
	APIKey         string  `mapstructure:"api_key"`
	EmbeddingModel string  `mapstructure:"embedding_model"`
	ChatModel      string  `mapstructure:"chat_model"`
	MaxTokens      int     `mapstructure:"max_tokens"`
	Temperature    float64 `mapstructure:"temperature"`
}

// SQLiteConfig holds SQLite database settings
type SQLiteConfig struct {
	Path              string        `mapstructure:"path"`
	MaxOpenConns      int           `mapstructure:"max_open_conns"`
	MaxIdleConns      int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime   time.Duration `mapstructure:"conn_max_lifetime"`
	EnableWAL         bool          `mapstructure:"enable_wal"`
	EnableForeignKeys bool          `mapstructure:"enable_foreign_keys"`
}

// QdrantConfig holds Qdrant vector database settings
type QdrantConfig struct {
	URL            string `mapstructure:"url"`
	APIKey         string `mapstructure:"api_key"`
	CollectionName string `mapstructure:"collection_name"`
	VectorSize     int    `mapstructure:"vector_size"`
	Distance       string `mapstructure:"distance"`
}

// LoggingConfig holds logging settings
type LoggingConfig struct {
	Level    string `mapstructure:"level"`
	Format   string `mapstructure:"format"`
	FilePath string `mapstructure:"file_path"`
}
