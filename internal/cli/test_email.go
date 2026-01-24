package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/M1ngdaXie/go-local-rag-email/internal/domain"
	"github.com/M1ngdaXie/go-local-rag-email/internal/repository/email"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var testEmailCmd = &cobra.Command{
	Use:   "test-email",
	Short: "Test email repository (create, get, list emails)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// TODO: Step 1 - Create the email repository
		repo := email.NewSQLiteRepository(application.SQLiteDB(), application.Logger())

		// TODO: Step 2 - Create a test email
		testEmail := &domain.Email{
			// ID 必须唯一，使用 UUID 生成
			ID:       uuid.New().String(),
			// 对于单封邮件，ThreadID 通常等于 ID，或者是属于某个会话的 ID
			ThreadID: uuid.New().String(), 
			
			Subject:  "Test Email from CLI 🚀",
			From:     "me@example.com",
			Date:     time.Now(),
			Snippet:  "This is a generated test email to verify SQLite storage...",
			BodyText: "Hello! If you can see this, the persistence layer is working correctly.",
		}

		// TODO: Step 3 - Save the email to database
				fmt.Println("Creating test email...")
		      if err := repo.Create(ctx, testEmail); err != nil {
		          return err
		      }
		      fmt.Printf("✓ Created email: %s\n", testEmail.ID)

		// TODO: Step 4 - Retrieve the email by ID
		fmt.Println("\nRetrieving email...")
		      retrieved, err := repo.Get(ctx, testEmail.ID)
		      if err != nil {
		          return err
		      }
		      fmt.Printf("✓ Retrieved: %s - %s\n", retrieved.Subject, retrieved.From)

		// TODO: Step 5 - List all emails
		fmt.Println("\nListing all emails...")
		      emails, err := repo.List(ctx, email.Filter{}, email.Pagination{Limit: 10})
		      if err != nil {
		          return err
		      }
		      fmt.Printf("✓ Found %d emails\n", len(emails))
		      for i, e := range emails {
		          fmt.Printf("  %d. %s - %s\n", i+1, e.Subject, e.From)
		       }

		fmt.Println("\n✅ All tests passed!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(testEmailCmd)
}
