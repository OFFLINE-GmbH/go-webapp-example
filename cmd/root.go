package cmd

import (
	"fmt"
	"os"

	"go-webapp-example/internal/app"
	"go-webapp-example/pkg/log"

	"github.com/spf13/cobra"

	// Enable file support for golang-migrate.
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var rootCmd = &cobra.Command{
	Use:   "go-webapp-example",
	Short: "This is the main command",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Please provide an argument. Use help to get more information.")
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func boot() *app.Kernel {
	kernel, err := app.New()
	if err != nil {
		logger := log.New(os.Stderr, "debug", "")
		logger.Fatalf("failed to boot kernel: %s", err)
		os.Exit(2)
	}
	return kernel
}
