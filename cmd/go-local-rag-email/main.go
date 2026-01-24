package main

import (
	"fmt"
	"os"

	"github.com/M1ngdaXie/go-local-rag-email/internal/app"
	"github.com/M1ngdaXie/go-local-rag-email/internal/cli"
)

func main() {
	// Initialize application container
	application, err := app.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize application: %v\n", err)
		os.Exit(1)
	}
	defer application.Shutdown()

	if err := cli.Execute(application); err != nil {
	          fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	          os.Exit(1)
	      }
}
