package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/M1ngdaXie/go-local-rag-email/internal/repository/vector"
	"github.com/M1ngdaXie/go-local-rag-email/internal/service/llm"
	"github.com/M1ngdaXie/go-local-rag-email/internal/service/rag"
	"github.com/spf13/cobra"
)

func NewSearchCmd() *cobra.Command {
	var limit int
	var minScore float32

	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Semantic search over your emails",
		Long: `Search your emails using natural language.

Examples:
  email search "meeting notes from John"
  email search "invoices" --limit 10
  email search "budget discussions" --min-score 0.6`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := application.Config()
			log := application.Logger()

			vectorRepo := vector.NewQdrantRepository(application.QdrantClient(), cfg.Qdrant, log)

			llmSvc, err := llm.New(cfg.OpenAI)
			if err != nil {
				return fmt.Errorf("failed to create LLM service: %w", err)
			}

			ragSvc := rag.New(vectorRepo, llmSvc, log)

			// Step 2: Build query from args
			query := strings.Join(args, " ")
			if strings.TrimSpace(query) == "" {
				return fmt.Errorf("query cannot be empty")
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			// Step 4: Execute search
			results, err := ragSvc.Search(ctx, query, limit)
			if err != nil {
				return fmt.Errorf("search failed: %w", err)
			}

			
			// Hint: Loop through results, keep only those with Score >= minScore
			if minScore > 0 {
                filtered := make([]rag.SearchResult, 0, len(results))
                for _, r := range results {
                    if r.Score >= minScore {
                        filtered = append(filtered, r)
                    }
                }
                results = filtered
            }

			if len(results) == 0 {
				fmt.Println("No emails found matching your query. Try broader search terms.")
				return nil
			}

			printSearchResults(results)

			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "n", 5, "Maximum number of results")
	cmd.Flags().Float32VarP(&minScore, "min-score", "s", 0.0, "Minimum relevance score (0.0-1.0)")

	return cmd
}

// printSearchResults formats and displays search results
func printSearchResults(results []rag.SearchResult) {
    w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0) // 间距调大一点点
    fmt.Fprintln(w, "SCORE\tFROM\tSUBJECT")
    fmt.Fprintln(w, "-----\t----\t-------")

    for _, r := range results {
        scoreStr := fmt.Sprintf("%.2f", r.Score)
        
        // 简单的条件颜色逻辑
        if r.Score >= 0.5 {
            scoreStr = "\033[32m" + scoreStr + "\033[0m" // 绿色
        } else if r.Score >= 0.3 {
            scoreStr = "\033[33m" + scoreStr + "\033[0m" // 黄色
        }

        // 还可以给 From 字段做一点脱敏或简化处理
        from := truncate(r.From, 20)
        
        fmt.Fprintf(w, "%s\t%s\t%s\n", scoreStr, from, truncate(r.Subject, 60))
    }

    w.Flush()
    fmt.Printf("\nFound %d relevant emails.\n", len(results))
}

// truncate shortens a string to maxLen, adding "..." if truncated
func truncate(s string, maxLen int) string {
    runes := []rune(s) // 转换为 rune 切片处理多字节字符
    if len(runes) <= maxLen {
        return s
    }
    return string(runes[:maxLen-3]) + "..."
}

func init() {
    rootCmd.AddCommand(NewSearchCmd())
}