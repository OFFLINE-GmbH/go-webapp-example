package quote

import (
	"context"
	"database/sql"

	"go-webapp-example/internal/pkg/audit"
	"go-webapp-example/internal/pkg/entity"
	"go-webapp-example/pkg/clock"
	"go-webapp-example/pkg/db"
	"go-webapp-example/pkg/util"

	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
)

// Store handles the direct database access for this entity.
type Store struct {
	db      *db.Connection
	clock   *clock.Clock
	auditor audit.ChangeAuditor
}

// NewStore returns a new store instance.
func NewStore(conn *db.Connection, auditor audit.ChangeAuditor, opts ...func(s *Store)) *Store {
	s := &Store{db: conn, auditor: auditor}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Find finds the entity by id.
func (s Store) Find(ctx context.Context, id int) (*entity.Quote, error) {
	var quote entity.Quote
	err := s.db.GetContext(ctx, &quote, "SELECT * FROM quotes WHERE id = ? LIMIT 1", id)
	return &quote, errors.WithStack(checkNotFound(err))
}

// Get returns all available entities.
func (s Store) Get(ctx context.Context) ([]*entity.Quote, error) {
	var quotes []*entity.Quote
	err := s.db.SelectContext(ctx, &quotes, "SELECT * FROM quotes")
	return quotes, errors.WithStack(err)
}

// GetByID returns calltypes by ID.
func (s Store) GetByID(ctx context.Context, ids []int) (map[int]*entity.Quote, error) {
	var calltypes []*entity.Quote
	query, params, err := sq.Select("*").From("quotes").Where(sq.Eq{"id": util.UniqueInts(ids)}).ToSql()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	err = s.db.SelectContext(ctx, &calltypes, query, params...)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	result := make(map[int]*entity.Quote, len(calltypes))
	for _, item := range calltypes {
		result[item.ID] = item
	}
	return result, errors.WithStack(err)
}

// Create creates a new entity.
func (s Store) Create(ctx context.Context, quote *entity.Quote) (*entity.Quote, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return quote, errors.WithStack(err)
	}
	quote.CreatedAt = s.clock.Now()
	quote.UpdatedAt = s.clock.Now()
	query, params, err := sq.Insert("quotes").SetMap(mapCols(quote)).ToSql()
	if err != nil {
		return quote, errors.WithStack(err)
	}

	res, err := tx.ExecContext(ctx, query, params...)
	if err != nil {
		return quote, db.RollbackError(tx, errors.WithStack(err))
	}
	id, err := res.LastInsertId()
	if err != nil {
		return quote, db.RollbackError(tx, errors.WithStack(err))
	}
	quote.ID = int(id)
	err = s.auditor.LogCreate(ctx, tx, quote)
	if err != nil {
		return quote, db.RollbackError(tx, errors.WithStack(err))
	}
	return quote, errors.WithStack(tx.Commit())
}

// Update saves an updated entity to the database.
func (s Store) Update(ctx context.Context, quote *entity.Quote) (*entity.Quote, error) {
	current, err := s.Find(ctx, quote.ID)
	if err != nil {
		return quote, errors.WithStack(err)
	}

	quote.CreatedAt = current.CreatedAt
	quote.UpdatedAt = s.clock.Now()

	query, params, err := sq.Update("quotes").SetMap(mapCols(quote)).Where(sq.Eq{"id": quote.ID}).ToSql()
	if err != nil {
		return quote, errors.WithStack(err)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return quote, errors.WithStack(err)
	}
	_, err = tx.ExecContext(ctx, query, params...)
	if err != nil {
		return quote, db.RollbackError(tx, errors.WithStack(err))
	}

	err = s.auditor.LogUpdate(ctx, tx, current, quote)
	if err != nil {
		return quote, db.RollbackError(tx, errors.WithStack(err))
	}

	return quote, errors.WithStack(tx.Commit())
}

// Delete removes multiple entities from the database.
func (s Store) Delete(ctx context.Context, tx *db.Tx, ids []int) ([]*entity.Quote, error) {
	var result []*entity.Quote
	sources, err := s.GetByID(ctx, ids)
	if err != nil {
		return result, errors.WithStack(err)
	}
	if len(sources) < 1 {
		return result, nil
	}
	for _, source := range sources {
		if r, err := s.deleteEntity(ctx, tx, source); err == nil {
			result = append(result, r)
		} else {
			return result, errors.WithStack(err)
		}
	}
	return result, nil
}

// deleteEntity removes a single entity from the database.
func (s Store) deleteEntity(ctx context.Context, tx *db.Tx, source *entity.Quote) (*entity.Quote, error) {
	_, err := tx.ExecContext(ctx, "DELETE FROM quotes WHERE id = ? LIMIT 1", source.ID)
	if err != nil {
		return source, errors.WithStack(err)
	}
	err = s.auditor.LogDelete(ctx, tx, source)
	if err != nil {
		return source, errors.WithStack(err)
	}
	return source, nil
}

// mapCols maps the entity to all default columns.
func mapCols(quote *entity.Quote) db.ColumnMap {
	return db.ColumnMap{
		"id":         quote.ID,
		"author":     quote.Author,
		"content":    quote.Content,
		"updated_at": quote.UpdatedAt,
		"created_at": quote.CreatedAt,
	}
}

// checkNotFound returns a ErrNotFound if no rows were returned.
func checkNotFound(err error) error {
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	return err
}
