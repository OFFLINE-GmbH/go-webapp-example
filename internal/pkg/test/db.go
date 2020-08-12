package test

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"go-webapp-example/pkg/db"
	"go-webapp-example/pkg/log"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/DATA-DOG/go-txdb"
	"github.com/romanyx/polluter"

	// enable mysql support for testing
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	// enable file source fo golang-migrate
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// dsn is the connection string used for these tests.
var dsn string

// nolint:gochecknoinits
func init() {
	host := "localhost"
	if os.Getenv("CI") != "" {
		host = "mysql"
	}
	dsn = fmt.Sprintf(
		"%s:%s@(%s:%d)/%s?multiStatements=true&parseTime=true&collation=utf8mb4_general_ci",
		"gowebapp",
		"gowebapp",
		host,
		3306,
		"gowebapp",
	)
	txdb.Register("txdb", "mysql", dsn)
}

// MockDB returns a mock DB connection.
func MockDB(t *testing.T) (*db.Connection, sqlmock.Sqlmock) {
	logger := log.New(os.Stdout, "error", "")

	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %s", err)
	}

	dbx := db.NewFromConnection(sqlx.NewDb(mockDB, "mysql"), logger)

	return dbx, mock
}

// DB returns a DB connection for the test database.
func DB(t *testing.T) (dbConn *db.Connection, cleanup func()) {
	logger := log.NewNullLogger()

	// Create a "proper" connection to run the migrations.
	dbNormal, err := sqlx.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("cannot connect to database: %s", err)
	}
	dbConn = db.NewFromConnection(dbNormal, logger)

	// Migrate
	migrationsPath := fmt.Sprintf("file://%s/../../../deployments/migrations", getBasePath())
	m, err := db.NewMigrator(dbConn, logger, migrationsPath)
	if err != nil {
		t.Fatalf("failed to create migrator: %s", err)
	}
	if err = m.Fresh(); err != nil {
		t.Fatalf("failed to migrate: %s", err)
	}

	// Seed
	seed, err := os.Open(filepath.Join(getBasePath(), "..", "..", "..", "deployments", "seeds", "testing.yml"))
	if err != nil {
		t.Fatalf("failed to open seed file: %s", err)
	}
	defer seed.Close()

	p := polluter.New(polluter.MySQLEngine(dbConn.Connection()))
	if err = p.Pollute(seed); err != nil {
		t.Fatalf("failed to pollute: %s", err)
	}

	// Connect
	return dbConn, func() {
		if err := dbConn.Close(); err != nil {
			t.Fatalf("db cleanup failed: %s", err)
		}
	}
}

// getBasePath returns the current base path of the process.
// nolint:dogsled
func getBasePath() string {
	_, b, _, _ := runtime.Caller(0)
	return filepath.Dir(b)
}
