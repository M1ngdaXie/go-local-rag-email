package cli

import (
	"fmt"
	"strings"

	"github.com/M1ngdaXie/go-local-rag-email/internal/service/gmail"
	"github.com/spf13/cobra"
)

var testParserCmd = &cobra.Command{
	Use:   "test-parser",
	Short: "Test email parser with a filthy HTML email",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("=== Testing Email Parser with Filthy HTML ===\n")

		// 1. Get the test email
		msg := gmail.GetFilthyEmail()
		fmt.Printf("Raw Message ID: %s\n", msg.Id)
		fmt.Printf("InternalDate (ms): %d\n\n", msg.InternalDate)

		// 2. Parse it
		parsed := gmail.ParseMessage(msg)

		// 3. Output results
		fmt.Println("--- Parsed Results ---")
		fmt.Printf("ID:       %s\n", parsed.ID)
		fmt.Printf("From:     %s\n", parsed.From)
		fmt.Printf("Subject:  %s\n", parsed.Subject)
		fmt.Printf("Date:     %s\n", parsed.Date.Format("2006-01-02 15:04:05 -0700"))
		fmt.Printf("Snippet:  %s\n", parsed.Snippet)
		fmt.Println()
		fmt.Println("--- Body Text (Extracted from HTML) ---")
		fmt.Println(parsed.BodyText)
		fmt.Println()

		// 4. Validation checks
		fmt.Println("--- Validation ---")

		// Check date is not zero
		if parsed.Date.IsZero() {
			fmt.Println("FAIL: Date is zero value (0001-01-01)")
		} else {
			fmt.Println("PASS: Date parsed correctly")
		}

		// Check body is not empty
		if parsed.BodyText == "" {
			fmt.Println("FAIL: Body text is empty")
		} else {
			fmt.Println("PASS: Body text extracted")
		}

		// Check HTML tags are stripped
		if strings.Contains(parsed.BodyText, "<") || strings.Contains(parsed.BodyText, ">") {
			fmt.Println("WARN: Body may still contain HTML tags")
		} else {
			fmt.Println("PASS: HTML tags stripped")
		}

		// Check entities are decoded
		if strings.Contains(parsed.BodyText, "&amp;") || strings.Contains(parsed.BodyText, "&nbsp;") {
			fmt.Println("WARN: HTML entities may not be fully decoded")
		} else {
			fmt.Println("PASS: HTML entities decoded")
		}

		fmt.Println("\n=== Test Complete ===")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(testParserCmd)
}
