package cmd

import (
	"context"

	"go-webapp-example/pkg/db"
	"go-webapp-example/pkg/log"

	"github.com/spf13/cobra"
)

const ActionUp = "up"
const ActionDown = "down"
const ActionFresh = "fresh"
const ActionVersion = "version"

// nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(migrateCmd)
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate the database",
	Long:  `Recreate the whole database`,
	Args:  cobra.MaximumNArgs(1),
	Run:   runMigrate,
}

func runMigrate(_ *cobra.Command, args []string) {
	app := boot()
	// nolint:errcheck
	defer app.Shutdown(context.Background())

	logger := app.Log.WithPrefix("cmd.migrate")

	m, err := db.NewMigrator(app.DB, logger, app.Config.Database.Migrations)
	if err != nil {
		logger.Fatalf("failed to create db migrator: %s", err)
	}

	action := ActionUp
	if len(args) == 1 {
		action = args[0]
	}

	handleAction(action, m, logger)
}

func handleAction(action string, m *db.Migrator, logger log.Logger) {
	switch action {
	case ActionUp:
		if err := m.Up(); err != nil {
			logger.Fatalf("failed to migrate up: %s", err)
		}
	case ActionDown:
		if err := m.Down(); err != nil {
			logger.Fatalf("failed to migrate down: %s", err)
		}
	case ActionFresh:
		if err := m.Fresh(); err != nil {
			logger.Fatalf("failed to refresh migrations: %s", err)
		}
	case ActionVersion:
		version, dirty, err := m.Version()
		if err != nil {
			logger.Fatalf("failed to get migration version: %s", err)
		}
		logger.
			WithFields(log.Fields{"version": version, "dirty": dirty}).
			Info("Current migration version")
	default:
		logger.Fatalf("unknown migration action: %s", action)
	}
}
