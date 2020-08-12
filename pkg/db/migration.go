package db

import (
	"fmt"
	"strings"

	"go-webapp-example/pkg/log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/pkg/errors"
)

type Migrator struct {
	db         *Connection
	migrate    *migrate.Migrate
	log        log.Logger
	migrations string
}

func NewMigrator(db *Connection, logger log.Logger, migrationsPath string) (*Migrator, error) {
	var err error
	var driver database.Driver

	// Add the scheme to the migrations path if not present.
	if !strings.HasPrefix(migrationsPath, "file://") {
		migrationsPath = fmt.Sprintf("file://%s", migrationsPath)
	}

	logger.WithFields(log.Fields{"migrations": migrationsPath, "driver": db.DriverName()}).Debugln("creating db migrator")

	driver, err = mysql.WithInstance(db.Connection(), &mysql.Config{})

	if err != nil {
		return &Migrator{}, errors.Wrap(err, "migrator: failed to generate database driver")
	}

	m, err := migrate.NewWithDatabaseInstance(migrationsPath, db.DriverName(), driver)
	if err != nil {
		return &Migrator{}, errors.Wrap(err, "migrator: failed to connect to database")
	}

	m.Log = logger

	return &Migrator{
		db:         db,
		migrate:    m,
		log:        logger,
		migrations: migrationsPath,
	}, nil
}

func (m *Migrator) Up() error {
	if err := m.migrate.Up(); err != nil {
		if err == migrate.ErrNoChange {
			m.log.Infoln("nothing to migrate")
			return nil
		}
		return err
	}
	return nil
}

func (m *Migrator) Down() error {
	if err := m.migrate.Down(); err != nil {
		if err == migrate.ErrNoChange {
			m.log.Infoln("nothing to migrate")
			return nil
		}
		return err
	}
	return nil
}

// Fresh drops the current database structure and re-runs all migrations.
func (m *Migrator) Fresh() error {
	err := m.migrate.Drop()
	if err != nil {
		return errors.Wrap(err, "failed to rollback migrations")
	}

	// Recreate migrator instance to get the schema_migrations table back.
	// @see https://github.com/golang-migrate/migrate/issues/226
	newInstance, err := NewMigrator(m.db, m.log, m.migrations)
	if err != nil {
		return errors.Wrap(err, "failed to create new migrator instance")
	}

	m.migrate = newInstance.migrate

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return errors.Wrap(err, "failed to run migrations")
	}
	return nil
}

func (m *Migrator) Version() (version uint, dirty bool, err error) {
	return m.migrate.Version()
}
