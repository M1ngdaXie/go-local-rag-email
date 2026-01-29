package cli

import (
	"context"
	"fmt"

	"github.com/M1ngdaXie/go-local-rag-email/internal/service/llm"
	"github.com/spf13/cobra"
)

var testLLMCmd = &cobra.Command{
	Use:   "test-llm",
	Short: "Test LLM service (OpenAI embeddings)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		fmt.Println("=== Testing LLM Service (OpenAI Embeddings) ===\n")

		// 1. Create LLM service using config
		cfg := application.Config()
		fmt.Printf("Model: %s\n", cfg.OpenAI.EmbeddingModel)

		svc, err := llm.New(cfg.OpenAI)
		if err != nil {
			return fmt.Errorf("failed to create LLM service: %w", err)
		}
		fmt.Println("LLM service created successfully\n")

		// 2. Test single embedding
		fmt.Println("--- Test 1: Single Embedding ---")
		testText := "This is a test email about a job interview at Google."
		fmt.Printf("Input text: %q\n", testText)

		embedding, err := svc.GenerateEmbedding(ctx, testText)
		if err != nil {
			fmt.Printf("FAIL: %v\n", err)
		} else {
			fmt.Printf("PASS: Generated embedding with %d dimensions\n", len(embedding))
			if len(embedding) > 0 {
				fmt.Printf("First 5 values: %v\n", embedding[:min(5, len(embedding))])
			}
		}

		// 3. Test batch embeddings
		fmt.Println("\n--- Test 2: Batch Embeddings ---")
		testTexts := []string{
			"Meeting scheduled for tomorrow at 2pm.",
			"Your Amazon order has shipped.",
			"Reminder: dentist appointment on Friday.",
		}
		fmt.Printf("Input: %d texts\n", len(testTexts))

		embeddings, err := svc.GenerateEmbeddings(ctx, testTexts)
		if err != nil {
			fmt.Printf("FAIL: %v\n", err)
		} else {
			fmt.Printf("PASS: Generated %d embeddings\n", len(embeddings))
			for i, emb := range embeddings {
				fmt.Printf("  [%d] %d dimensions\n", i, len(emb))
			}
		}

		// 4. Test empty input handling
		fmt.Println("\n--- Test 3: Empty Input Handling ---")
		_, err = svc.GenerateEmbedding(ctx, "")
		if err != nil {
			fmt.Printf("PASS: Empty input correctly rejected: %v\n", err)
		} else {
			fmt.Println("WARN: Empty input was accepted (should be rejected)")
		}

		// 5. Validation summary
		fmt.Println("\n--- Validation ---")
		if embedding != nil && len(embedding) == 1536 {
			fmt.Println("PASS: Vector dimension is 1536 (matches text-embedding-3-small)")
		} else if embedding != nil {
			fmt.Printf("INFO: Vector dimension is %d\n", len(embedding))
		}

		fmt.Println("\n=== Test Complete ===")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(testLLMCmd)
}
