package cmd

import (
	"fmt"

	"go-webapp-example/internal/app"

	"github.com/spf13/cobra"
)

// nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of the binary",
	Long:  "Print the version number of the binary",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(app.Version())
	},
}
