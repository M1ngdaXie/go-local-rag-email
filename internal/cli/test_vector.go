package cli

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/M1ngdaXie/go-local-rag-email/internal/repository/vector"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var testVectorCmd = &cobra.Command{
	Use:   "test-vector",
	Short: "Test Qdrant vector operations (upsert, search, delete)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Step 1: Create vector repository
		repo := vector.NewQdrantRepository(
			application.QdrantClient(),
			application.Config().Qdrant,
			application.Logger(),
		)

		// Step 2: Check collection info
		fmt.Println("📊 Checking collection info...")
		info, err := repo.CollectionInfo(ctx)
		if err != nil {
			return fmt.Errorf("failed to get collection info: %w", err)
		}
		fmt.Printf("✓ Collection: %d vectors, %d points, Status: %s\n",
			info.VectorsCount, info.PointsCount, info.Status)

		// Step 3: Create test vectors (fake embeddings)
		fmt.Println("\n📝 Creating test vectors...")
		testVectors := []*vector.Point{
			{
				ID:     uuid.New().String(),
				Vector: generateFakeVector(1536), // 1536 dimensions for text-embedding-3-small
				Payload: map[string]interface{}{
					"email_id": "test-email-1",
					"chunk_id": "chunk-1",
					"content":  "This is a test email about machine learning and AI",
				},
			},
			{
				ID:     uuid.New().String(),
				Vector: generateFakeVector(1536),
				Payload: map[string]interface{}{
					"email_id": "test-email-2",
					"chunk_id": "chunk-2",
					"content":  "Another email about deep learning, neural networks, and transformers",
				},
			},
			{
				ID:     uuid.New().String(),
				Vector: generateFakeVector(1536),
				Payload: map[string]interface{}{
					"email_id": "test-email-3",
					"chunk_id": "chunk-3",
					"content":  "Email discussing Python programming and software engineering",
				},
			},
		}

		// Step 4: Upsert vectors to Qdrant
		fmt.Println("\n⬆️  Upserting test vectors...")
		if err := repo.Upsert(ctx, testVectors); err != nil {
			return fmt.Errorf("failed to upsert: %w", err)
		}
		fmt.Printf("✓ Upserted %d vectors\n", len(testVectors))

		// Step 5: Search for similar vectors
		fmt.Println("\n🔍 Searching for similar vectors...")
		// Use the first vector as query (simulating a search)
		results, err := repo.Search(ctx, testVectors[0].Vector, vector.SearchOptions{
			Limit:          5,
			ScoreThreshold: 0.0,
		})
		if err != nil {
			return fmt.Errorf("failed to search: %w", err)
		}

		fmt.Printf("✓ Found %d results:\n", len(results))
		for i, r := range results {
			content := r.Payload["content"].(string)
			emailID := r.Payload["email_id"].(string)
			fmt.Printf("  %d. Score: %.4f | Email: %s\n     Content: %s\n",
				i+1, r.Score, emailID, content)
		}

		// Step 6: Test DeleteByEmailID (payload filtering)
		fmt.Println("\n🗑️  Testing delete by email_id...")
		if err := repo.DeleteByEmailID(ctx, "test-email-2"); err != nil {
			return fmt.Errorf("failed to delete by email_id: %w", err)
		}
		fmt.Println("✓ Deleted all vectors for email 'test-email-2'")

		// Step 7: Delete remaining vectors by ID
		fmt.Println("\n🗑️  Cleaning up remaining test vectors...")
		ids := []string{testVectors[0].ID, testVectors[2].ID}
		if err := repo.Delete(ctx, ids); err != nil {
			return fmt.Errorf("failed to delete: %w", err)
		}
		fmt.Printf("✓ Deleted %d vectors\n", len(ids))

		// Step 8: Verify collection is clean
		fmt.Println("\n📊 Final collection info...")
		finalInfo, err := repo.CollectionInfo(ctx)
		if err != nil {
			return fmt.Errorf("failed to get final info: %w", err)
		}
		fmt.Printf("✓ Collection: %d vectors remaining\n", finalInfo.VectorsCount)

		fmt.Println("\n✅ All vector tests passed!")
		return nil
	},
}

// generateFakeVector creates a random vector for testing
// In production, vectors come from OpenAI's embedding API
func generateFakeVector(size int) []float32 {
	vec := make([]float32, size)
	for i := 0; i < size; i++ {
		vec[i] = rand.Float32() // Random value between 0.0 and 1.0
	}
	return vec
}

func init() {
	rootCmd.AddCommand(testVectorCmd)
}
