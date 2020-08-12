package db

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go-webapp-example/pkg/log"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gopkg.in/guregu/null.v3"
)

// ErrNotExists is returned when a query is executed on a non-existing entity.
var ErrNotExists = errors.New("cannot update a non-existing entity")

// Regex used to convert field names to snake case.
var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

// ColumnMap is a map used to map column to their values.
type ColumnMap map[string]interface{}

const TypeMySQL = "mysql"

type Connection struct {
	*sqlx.DB
	log log.Logger
}

type Rows struct {
	*sqlx.Rows
}

type Row struct {
	*sqlx.Row
}

type Tx struct {
	*sqlx.Tx
	Log log.Logger
}

type Result struct {
	sql.Result
}

// New returns a new Connection instance.
func New(driver, dataSource string, logger log.Logger) (*Connection, error) {
	logger.WithFields(log.Fields{"driver": driver, "source": dataSource}).Traceln("creating db connection")

	db, err := sqlx.Connect(driver, dataSource)
	if err != nil {
		return nil, err
	}
	return &Connection{DB: configure(db), log: logger}, err
}

// NewFromConnection returns a new Connection for an existing sqlx instance.
func NewFromConnection(conn *sqlx.DB, logger log.Logger) *Connection {
	logger.WithFields(log.Fields{"driver": conn.DriverName()}).Traceln("reusing db connection")
	return &Connection{DB: configure(conn), log: logger}
}

// configure sets default connection settings.
func configure(db *sqlx.DB) *sqlx.DB {
	db.SetConnMaxLifetime(time.Second)
	db.MapperFunc(toSnakeCase)
	return db
}

func (c *Connection) Close() error {
	c.log.Traceln("closing db connection")
	return c.DB.Close()
}

func (c *Connection) Begin() (*Tx, error) {
	tx, err := c.DB.Beginx()
	defer logQueryWithArgs(c.log, time.Now(), "BEGIN TRANSACTION", nil)
	if err != nil {
		defer logErrorWithArgs(c.log, "BEGIN TRANSACTION", nil, err)
	}
	return &Tx{Tx: tx, Log: c.log}, err
}

func (c *Connection) Query(query string, args ...interface{}) (*Rows, error) {
	defer logQueryWithArgs(c.log, time.Now(), query, args)
	rows, err := c.DB.Queryx(query, args...)
	return &Rows{rows}, err
}

func (c *Connection) QueryContext(ctx context.Context, query string, args ...interface{}) (*Rows, error) {
	defer logQueryWithArgs(c.log, time.Now(), query, args)
	rows, err := c.DB.QueryxContext(ctx, query, args...)
	return &Rows{rows}, err
}

func (c *Connection) Select(dest interface{}, query string, args ...interface{}) error {
	defer logQueryWithArgs(c.log, time.Now(), query, args)
	err := c.DB.Select(dest, query, args...)
	if err != nil && err.Error() != "sql: no rows in result set" {
		defer logErrorWithArgs(c.log, query, args, err)
	}
	return err
}

func (c *Connection) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	defer logQueryWithArgs(c.log, time.Now(), query, args)
	err := c.DB.SelectContext(ctx, dest, query, args...)
	if err != nil && err.Error() != "sql: no rows in result set" {
		defer logErrorWithArgs(c.log, query, args, err)
	}
	return err
}

func (c *Connection) Exec(query string, args ...interface{}) (Result, error) {
	defer logQueryWithArgs(c.log, time.Now(), query, args)
	res, err := c.DB.Exec(query, args...)
	if err != nil {
		defer logErrorWithArgs(c.log, query, args, err)
	}
	return Result{res}, err
}

func (c *Connection) ExecContext(ctx context.Context, query string, args ...interface{}) (Result, error) {
	defer logQueryWithArgs(c.log, time.Now(), query, args)
	res, err := c.DB.ExecContext(ctx, query, args...)
	if err != nil {
		defer logErrorWithArgs(c.log, query, args, err)
	}
	return Result{res}, err
}

func (c *Connection) Get(dest interface{}, query string, args ...interface{}) error {
	defer logQueryWithArgs(c.log, time.Now(), query, args)
	err := c.DB.Get(dest, query, args...)
	if err != nil && err.Error() != "sql: no rows in result set" {
		defer logErrorWithArgs(c.log, query, args, err)
	}
	return err
}

func (c *Connection) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	defer logQueryWithArgs(c.log, time.Now(), query, args)
	err := c.DB.GetContext(ctx, dest, query, args...)
	if err != nil && err.Error() != "sql: no rows in result set" {
		defer logErrorWithArgs(c.log, query, args, err)
	}
	return err
}

func (c *Connection) Connection() *sql.DB {
	return c.DB.DB
}

func (c *Connection) DriverName() string {
	return c.DB.DriverName()
}

// WithTX runs a closure inside a transaction.
func (c *Connection) WithTx(fn func(tx *Tx) error) error {
	tx, err := c.Begin()
	if err != nil {
		return errors.WithStack(err)
	}
	err = fn(tx)
	if err != nil {
		return tx.Rollback()
	}
	return tx.Commit()
}

func (tx *Tx) Exec(query string, args ...interface{}) (Result, error) {
	defer logQueryWithArgs(tx.Log, time.Now(), query, args)
	r, err := tx.Tx.Exec(query, args...)
	if err != nil {
		defer logErrorWithArgs(tx.Log, query, args, err)
	}
	return Result{r}, err
}

func (tx *Tx) ExecContext(ctx context.Context, query string, args ...interface{}) (Result, error) {
	defer logQueryWithArgs(tx.Log, time.Now(), query, args)
	r, err := tx.Tx.ExecContext(ctx, query, args...)
	if err != nil {
		defer logErrorWithArgs(tx.Log, query, args, err)
	}
	return Result{r}, err
}

func (tx *Tx) Rollback() error {
	defer logQuery(tx.Log, time.Now(), "ROLLBACK")
	err := tx.Tx.Rollback()
	if err != nil {
		defer logErrorWithArgs(tx.Log, "ROLLBACK", nil, err)
	}
	return err
}

func (tx *Tx) Commit() error {
	defer logQuery(tx.Log, time.Now(), "COMMIT")
	err := tx.Tx.Commit()
	if err != nil {
		defer logErrorWithArgs(tx.Log, "ROLLBACK", nil, err)
	}
	return err
}

// logQueryWithArgs times and logs a executed query with arguments.
func logQueryWithArgs(logger log.Logger, start time.Time, query string, args []interface{}) {
	query = strings.ReplaceAll(query, "?", "%v")
	query = fmt.Sprintf(query, args...)

	logQuery(logger, start, query)
}

// logErrorWithArgs times and logs a executed query with arguments.
func logErrorWithArgs(logger log.Logger, query string, args []interface{}, err error) {
	query = strings.ReplaceAll(query, "?", "%v")
	query = fmt.Sprintf(query, args...)

	logger.WithFields(log.Fields{"error": err}).Errorf(query)
}

// logQuery times and logs a executed query.
func logQuery(logger log.Logger, start time.Time, query string) {
	duration := time.Since(start)

	// Log slow queries as warnings.
	if duration > 100*time.Millisecond || query == "ROLLBACK" {
		logger.WithFields(log.Fields{"time": duration.String()}).Warnln(query)
	} else {
		logger.WithFields(log.Fields{"time": duration.String()}).Traceln(query)
	}
}

// RollbackError rolls back a transaction and returns the provided error with a stack trace.
func RollbackError(tx *Tx, originalErr error) error {
	err := tx.Rollback()
	if err != nil {
		return errors.Wrap(err, "failed to rollback transaction")
	}
	return originalErr
}

// toSnakeCase converts a input string to snake_case.
func toSnakeCase(input string) string {
	snake := matchFirstCap.ReplaceAllString(input, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

// RawTime handles MySQL's TIME column type.
// @see https://github.com/go-sql-driver/mysql#timetime-support
type RawTime []byte

func (t RawTime) Time() (null.Time, error) {
	if len(t) == 0 {
		return null.Time{}, nil
	}
	parsed, err := time.Parse("15:04:05", string(t))
	if err != nil {
		return null.Time{}, errors.WithStack(err)
	}
	return null.TimeFrom(parsed), nil
}
