# Go-Local-RAG-Email

A local-first RAG (Retrieval-Augmented Generation) email assistant built in Go with CLI/TUI interface.

## Features

- **Email Sync**: Fetch emails from Gmail with OAuth 2.0
- **Semantic Search**: Natural language search using vector embeddings
- **AI Summarization**: GPT-4 powered email summaries
- **Interactive TUI**: Terminal UI with Bubbletea
- **Local-First**: All data stored locally (SQLite + Qdrant)

## Tech Stack

- **Language**: Go 1.22+
- **CLI Framework**: Cobra + Viper
- **TUI Framework**: Bubbletea + Bubbles + Lipgloss
- **Databases**: SQLite (metadata) + Qdrant (vector embeddings)
- **Email**: Gmail API with OAuth 2.0
- **AI**: OpenAI API (embeddings + chat)

## Prerequisites

- Go 1.22 or higher
- Docker and Docker Compose
- OpenAI API key
- Gmail API credentials (optional, for email sync)

## Quick Start

### 1. Clone the repository

```bash
git clone https://github.com/M1ngdaXie/go-local-rag-email.git
cd go-local-rag-email
```

### 2. Install dependencies

```bash
go mod download
```

### 3. Start Qdrant (vector database)

```bash
make docker-up
```

Qdrant will be available at `http://localhost:6333`

### 4. Configure the application

```bash
make setup
```

Then edit `~/.go-local-rag-email/config.yaml`:

```yaml
openai:
  api_key: "your-openai-api-key-here"

gmail:
  credentials_path: "configs/credentials.json"  # Add your Gmail API credentials
```

### 5. Build and run

```bash
make build
./bin/go-local-rag-email
```

Or run directly:

```bash
make run
```

## Project Structure

```
.
├── cmd/
│   └── go-local-rag-email/    # Application entry point
├── internal/
│   ├── app/                   # Dependency injection container
│   ├── cli/                   # Cobra CLI commands
│   ├── tui/                   # Bubbletea TUI interface
│   ├── domain/                # Domain models
│   ├── service/               # Business logic
│   │   ├── email/             # Email service (Gmail)
│   │   ├── rag/               # RAG pipeline
│   │   ├── llm/               # LLM service (OpenAI)
│   │   └── sync/              # Sync orchestration
│   ├── repository/            # Data access layer
│   │   ├── email/             # Email repository (SQLite)
│   │   └── vector/            # Vector repository (Qdrant)
│   ├── database/              # Database clients
│   └── config/                # Configuration management
├── pkg/                       # Reusable packages
│   ├── logger/
│   ├── errors/
│   ├── retry/
│   └── tokenstore/
├── configs/                   # Configuration files
├── scripts/                   # Helper scripts
└── test/                      # Tests
```

## Development

### Build

```bash
make build
```

### Run tests

```bash
make test
```

### Lint code

```bash
make lint
```

### Clean build artifacts

```bash
make clean
```

### Docker commands

```bash
make docker-up    # Start Qdrant
make docker-down  # Stop Qdrant
```

## Usage (Coming Soon)

Once implemented, the CLI will support:

```bash
# Sync emails from Gmail
go-local-rag-email sync --since 7d

# Search emails with natural language
go-local-rag-email search "quarterly budget review"

# Summarize an email
go-local-rag-email summarize <email-id>

# Launch interactive TUI
go-local-rag-email tui
```

## Configuration

Configuration is managed via `~/.go-local-rag-email/config.yaml`. See `configs/config.yaml.example` for all available options.

Key settings:
- **OpenAI API key**: Required for embeddings and chat
- **Gmail credentials**: Required for email sync
- **Qdrant URL**: Vector database endpoint (default: `http://localhost:6333`)
- **SQLite path**: Local database location

## Architecture

This project follows **Clean Architecture** principles:

- **Domain Layer**: Core business entities
- **Service Layer**: Business logic and orchestration
- **Repository Layer**: Data access abstraction
- **Interface Layer**: CLI/TUI presentation

Key patterns:
- Dependency Injection via central App container
- Interface-driven design for testability
- Standard Go Project Layout

## License

MIT License

## Contributing

This is a portfolio project. Feel free to fork and experiment!

## Roadmap

- [x] Project skeleton setup
- [ ] Configuration management with Viper
- [ ] Database layer (SQLite + Qdrant)
- [ ] Gmail OAuth flow
- [ ] Email sync service
- [ ] OpenAI integration (embeddings + chat)
- [ ] RAG pipeline (chunking + indexing)
- [ ] Semantic search
- [ ] CLI commands
- [ ] TUI interface
- [ ] Tests
