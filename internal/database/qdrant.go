package database

import (
	"context"
	"fmt"

	"github.com/M1ngdaXie/go-local-rag-email/internal/config"
	"github.com/M1ngdaXie/go-local-rag-email/pkg/logger"
	"github.com/qdrant/go-client/qdrant"
)

// NewQdrant creates a new Qdrant client connection
func NewQdrant(cfg config.QdrantConfig, log logger.Logger) (*qdrant.Client, error) {
	// Step 1: Create Qdrant client config
	clientConfig := &qdrant.Config{
		Host: cfg.URL,
	}

	// Add API key if provided (optional for local)
	if cfg.APIKey != "" {
		clientConfig.APIKey = cfg.APIKey
	}

	// Step 2: Create the client
	client, err := qdrant.NewClient(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Qdrant client: %w", err)
	}

	// Step 3: Test connection by listing collections
	ctx := context.Background()
	_, err = client.ListCollections(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Qdrant (is it running?): %w", err)
	}

	// Step 4: Log success
	log.Info("Qdrant client connected", "url", cfg.URL)

	// Step 5: Return the client
	return client, nil
}

// CreateEmailCollection creates the vector collection if it doesn't exist
func CreateEmailCollection(ctx context.Context, client *qdrant.Client, cfg config.QdrantConfig, log logger.Logger) error {
	// Step 1: Check if collection already exists
	exists, err := client.CollectionExists(ctx, cfg.CollectionName)
	if err != nil {
		return fmt.Errorf("failed to check collection existence: %w", err)
	}

	if exists {
		log.Info("Qdrant collection already exists", "name", cfg.CollectionName)
		return nil
	}

	// Step 2: Create the collection with vector configuration
	err = client.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: cfg.CollectionName,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     uint64(cfg.VectorSize),  // 1536 for text-embedding-3-small
			Distance: qdrant.Distance_Cosine,   // Cosine similarity for text
		}),
	})
	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	// Step 3: Log success
	log.Info("Created Qdrant collection", "name", cfg.CollectionName, "size", cfg.VectorSize)

	return nil
}
