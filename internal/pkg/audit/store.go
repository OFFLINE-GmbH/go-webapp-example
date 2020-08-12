package audit

import (
	"context"
	"database/sql"

	"go-webapp-example/internal/pkg/entity"
	"go-webapp-example/pkg/clock"
	"go-webapp-example/pkg/db"

	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"gopkg.in/guregu/null.v3"
)

// Store handles the direct database access for this entity.
type Store struct {
	db    *db.Connection
	clock *clock.Clock
}

// NewStore returns a new store instance.
func NewStore(conn *db.Connection, opts ...func(s *Store)) *Store {
	s := &Store{db: conn}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Find finds the entity by id.
func (s Store) Find(ctx context.Context, id int) (*entity.AuditLog, error) {
	var log entity.AuditLog
	err := s.db.GetContext(ctx, &log, "SELECT * FROM auditlogs WHERE id = ? LIMIT 1", id)
	return &log, errors.WithStack(checkNotFound(err))
}

// Get returns all available entities.
func (s Store) Get(ctx context.Context) ([]*entity.AuditLog, error) {
	var auditlogs []*entity.AuditLog
	err := s.db.SelectContext(ctx, &auditlogs, "SELECT * FROM auditlogs")
	return auditlogs, errors.WithStack(err)
}

// Create creates a new entity.
func (s Store) Create(ctx context.Context, tx *db.Tx, log *entity.AuditLog) (*entity.AuditLog, error) {
	tx, err := s.ensureTx(tx)
	if err != nil {
		return log, errors.WithStack(err)
	}
	log.CreatedAt = null.TimeFrom(s.clock.Now())
	log.UpdatedAt = null.TimeFrom(s.clock.Now())

	query, params, err := sq.Insert("auditlogs").SetMap(mapCols(log)).ToSql()
	if err != nil {
		return log, errors.WithStack(err)
	}

	res, err := tx.ExecContext(ctx, query, params...)
	if err != nil {
		return log, errors.WithStack(err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return log, errors.WithStack(err)
	}

	log.ID = int(id)
	return log, nil
}

// Update saves an updated entity to the database.
func (s Store) Update(ctx context.Context, tx *db.Tx, log *entity.AuditLog) (*entity.AuditLog, error) {
	if log.ID < 1 {
		return log, errors.WithStack(db.ErrNotExists)
	}
	tx, err := s.ensureTx(tx)
	if err != nil {
		return log, errors.WithStack(err)
	}

	log.UpdatedAt = null.TimeFrom(s.clock.Now())

	query, params, err := sq.Update("auditlogs").SetMap(mapCols(log)).Where(sq.Eq{"id": log.ID}).ToSql()
	if err != nil {
		return log, errors.WithStack(err)
	}

	_, err = tx.ExecContext(ctx, query, params...)
	return log, errors.WithStack(err)
}

// Delete removes an entity from the database.
func (s Store) Delete(ctx context.Context, tx *db.Tx, log *entity.AuditLog) (*entity.AuditLog, error) {
	if log.ID < 1 {
		return log, nil
	}
	tx, err := s.ensureTx(tx)
	if err != nil {
		return log, errors.WithStack(err)
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM auditlogs WHERE id = ? LIMIT 1", log.ID)
	return log, errors.WithStack(err)
}

// ensureTx makes sure a "nil" db.Tx is transformed into a proper Transaction.
func (s Store) ensureTx(tx *db.Tx) (*db.Tx, error) {
	if tx == nil {
		newTx, err := s.db.Begin()
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return newTx, nil
	}
	return tx, nil
}

// mapCols maps the entity to all default columns.
func mapCols(log *entity.AuditLog) db.ColumnMap {
	return db.ColumnMap{
		"action":      log.Action,
		"entity_id":   log.EntityID,
		"entity_type": log.EntityType,
		"field":       log.Field,
		"meta":        log.Meta,
		"user_id":     log.UserID,
		"value_new":   log.ValueNew,
		"value_old":   log.ValueOld,
		"created_at":  log.CreatedAt,
		"updated_at":  log.UpdatedAt,
	}
}

// checkNotFound returns a ErrNotFound if no rows were returned.
func checkNotFound(err error) error {
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	return err
}
