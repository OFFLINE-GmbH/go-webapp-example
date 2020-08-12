package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"go-webapp-example/pkg/db"

	"github.com/romanyx/polluter"
	"github.com/spf13/cobra"
)

// nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(seedCmd)
}

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seed data",
	Long:  `This command seeds example data into the database`,
	Run:   seed,
}

func seed(_ *cobra.Command, _ []string) {
	kernel := boot()

	logger := kernel.Log.WithPrefix("cmd.seed")

	m, err := db.NewMigrator(kernel.DB, logger, kernel.Config.Database.Migrations)
	if err != nil {
		logger.Fatalf("failed to create db migrator: %s", err)
	}

	err = m.Fresh()
	if err != nil {
		logger.Fatalf("failed to clean db: %s", err)
	}

	// Seed
	seed, err := os.Open(filepath.Join(getBasePath(), "../deployments/seeds/base.yml"))
	if err != nil {
		panic(fmt.Sprintf("failed to open seed file: %s", err))
	}
	defer seed.Close()

	p := polluter.New(polluter.MySQLEngine(kernel.DB.Connection()))
	if err = p.Pollute(seed); err != nil {
		panic(fmt.Sprintf("failed to pollute: %s", err))
	}
}

// getBasePath returns the current base path of the process.
// nolint:dogsled
func getBasePath() string {
	_, b, _, _ := runtime.Caller(0)
	return filepath.Dir(b)
}
