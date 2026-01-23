#!/bin/bash

set -e

echo "🚀 Setting up Go-Local-RAG-Email..."
echo ""

# Create data directory
echo "📁 Creating data directory..."
mkdir -p ~/.go-local-rag-email
echo "✓ Created ~/.go-local-rag-email"

# Copy config if doesn't exist
echo "⚙️  Setting up configuration..."
if [ ! -f ~/.go-local-rag-email/config.yaml ]; then
    cp configs/config.yaml.example ~/.go-local-rag-email/config.yaml
    echo "✓ Created config file at ~/.go-local-rag-email/config.yaml"
else
    echo "✓ Config file already exists"
fi

# Download Go dependencies
echo "📦 Downloading Go dependencies..."
go mod download
echo "✓ Dependencies downloaded"

# Start Docker services
echo "🐳 Starting Docker services..."
if command -v docker-compose &> /dev/null || command -v docker &> /dev/null; then
    if [ -f docker-compose.yml ]; then
        docker-compose up -d 2>/dev/null || docker compose up -d
        echo "✓ Qdrant started at http://localhost:6333"
    else
        echo "⚠️  docker-compose.yml not found, skipping Docker setup"
    fi
else
    echo "⚠️  Docker not found, skipping Docker setup"
    echo "   Install Docker to run Qdrant: https://docs.docker.com/get-docker/"
fi

echo ""
echo "✨ Setup complete!"
echo ""
echo "Next steps:"
echo "1. Edit ~/.go-local-rag-email/config.yaml"
echo "   - Add your OpenAI API key"
echo "   - Configure Gmail API credentials (optional)"
echo ""
echo "2. Build the application:"
echo "   make build"
echo ""
echo "3. Run the application:"
echo "   ./bin/go-local-rag-email"
echo ""
echo "For more information, see README.md"
