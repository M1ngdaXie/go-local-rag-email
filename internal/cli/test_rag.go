package cli

import (
	"context"
	"fmt"

	"github.com/M1ngdaXie/go-local-rag-email/internal/repository/email"
	"github.com/M1ngdaXie/go-local-rag-email/internal/repository/vector"
	"github.com/M1ngdaXie/go-local-rag-email/internal/service/llm"
	"github.com/M1ngdaXie/go-local-rag-email/internal/service/rag"
	"github.com/spf13/cobra"
)

var testRAGCmd = &cobra.Command{
	Use:   "test-rag",
	Short: "Test RAG pipeline (index emails and search)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		fmt.Println("=== Testing RAG Pipeline ===\n")

		// 1. Create dependencies
		cfg := application.Config()
		log := application.Logger()

		// Email repository (to fetch emails)
		emailRepo := email.NewSQLiteRepository(application.SQLiteDB(), log)

		// Vector repository (Qdrant)
		vectorRepo := vector.NewQdrantRepository(application.QdrantClient(), cfg.Qdrant, log)

		// LLM service (OpenAI embeddings)
		llmSvc, err := llm.New(cfg.OpenAI)
		if err != nil {
			return fmt.Errorf("failed to create LLM service: %w", err)
		}

		// RAG service
		ragSvc := rag.New(vectorRepo, llmSvc, log)

		fmt.Println("Services initialized successfully\n")

		// 2. Fetch emails from SQLite
		fmt.Println("--- Step 1: Fetching emails from SQLite ---")
		emails, err := emailRepo.List(ctx, email.Filter{}, email.Pagination{Limit: 50})
		if err != nil {
			return fmt.Errorf("failed to fetch emails: %w", err)
		}
		fmt.Printf("Found %d emails to index\n\n", len(emails))

		if len(emails) == 0 {
			fmt.Println("No emails found. Run 'sync' first to fetch emails from Gmail.")
			return nil
		}

		// 3. Index emails
		fmt.Println("--- Step 2: Indexing emails ---")
		err = ragSvc.IndexEmails(ctx, emails)
		if err != nil {
			fmt.Printf("Indexing error: %v\n", err)
		}

		// 4. Check collection info
		fmt.Println("\n--- Step 3: Collection Info ---")
		info, err := vectorRepo.CollectionInfo(ctx)
		if err != nil {
			fmt.Printf("Failed to get collection info: %v\n", err)
		} else {
			fmt.Printf("Vectors in collection: %d\n", info.VectorsCount)
			fmt.Printf("Points in collection: %d\n", info.PointsCount)
			fmt.Printf("Status: %s\n", info.Status)
		}

		// 5. Test search
		fmt.Println("\n--- Step 4: Test Search ---")
		testQuery := "Software engineer"
		fmt.Printf("Query: %q\n", testQuery)

		results, err := ragSvc.Search(ctx, testQuery, 5)
		if err != nil {
			fmt.Printf("Search error: %v\n", err)
		} else {
			fmt.Printf("Found %d results:\n", len(results))
			for i, r := range results {
				fmt.Printf("  %d. [%.2f] %s - %s\n", i+1, r.Score, r.Subject, r.From)
			}
		}

		fmt.Println("\n=== Test Complete ===")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(testRAGCmd)
}
