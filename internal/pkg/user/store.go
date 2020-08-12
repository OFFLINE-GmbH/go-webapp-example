package user

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
	"gopkg.in/guregu/null.v3"
)

type authManager interface {
	AddRoleForUser(userID int, roleID int) bool
	RemoveRoleForUser(userID int, roleID int) bool
}

// Store handles the direct database access for this entity.
type Store struct {
	db      *db.Connection
	clock   *clock.Clock
	auditor audit.ChangeAuditor
	auth    authManager
}

// NewStore returns a new store instance.
func NewStore(conn *db.Connection, auth authManager, auditor audit.ChangeAuditor, opts ...func(s *Store)) *Store {
	s := &Store{db: conn, auditor: auditor, auth: auth}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Find finds the entity by id.
func (s Store) Find(ctx context.Context, id int) (*entity.User, error) {
	var user entity.User
	err := s.db.GetContext(ctx, &user, "SELECT * FROM users WHERE id = ? LIMIT 1", id)
	return &user, errors.WithStack(checkNotFound(err))
}

// Find finds the entity by name.
func (s Store) FindByName(ctx context.Context, name string) (*entity.User, error) {
	var user entity.User
	err := s.db.GetContext(ctx, &user, "SELECT * FROM users WHERE name = ? LIMIT 1", name)
	return &user, errors.WithStack(checkNotFound(err))
}

// Get returns all available entities.
func (s Store) Get(ctx context.Context) ([]*entity.User, error) {
	var users []*entity.User
	err := s.db.SelectContext(ctx, &users, "SELECT * FROM users")
	return users, errors.WithStack(err)
}

// GetByID returns the entity by ID.
func (s Store) GetByID(ctx context.Context, ids []int) (map[int]*entity.User, error) {
	var buttons []*entity.User
	query, params, err := sq.Select("*").From("users").Where(sq.Eq{"id": util.UniqueInts(ids)}).ToSql()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	err = s.db.SelectContext(ctx, &buttons, query, params...)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	result := make(map[int]*entity.User, len(buttons))
	for _, item := range buttons {
		result[item.ID] = item
	}
	return result, errors.WithStack(err)
}

// GetByRoleID returns a map of role ids to a slice of permissions.
func (s Store) GetByRoleID(ctx context.Context, ids []int) (map[int][]*entity.User, error) {
	type result struct {
		entity.User
		RoleID int `db:"role_id"`
	}
	var users []*result
	ret := make(map[int][]*entity.User)
	builder := sq.
		Select("users.*, role_user.role_id as role_id").
		From("role_user").
		LeftJoin("users ON role_user.user_id = users.id")
	if len(ids) > 0 {
		builder = builder.Where(sq.Eq{"role_user.role_id": ids})
	}
	query, params, err := builder.ToSql()
	if err != nil {
		return ret, errors.WithStack(err)
	}
	err = s.db.SelectContext(ctx, &users, query, params...)
	if err != nil {
		return ret, errors.WithStack(err)
	}
	for _, user := range users {
		ret[user.RoleID] = append(ret[user.RoleID], &entity.User{
			ID:          user.ID,
			Name:        user.Name,
			Password:    user.Password,
			IsSuperuser: user.IsSuperuser,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		})
	}
	return ret, errors.WithStack(checkNotFound(err))
}

// Create creates a new entity.
func (s Store) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return user, errors.WithStack(err)
	}
	user.CreatedAt = null.TimeFrom(s.clock.Now())
	user.UpdatedAt = null.TimeFrom(s.clock.Now())
	res, err := tx.ExecContext(ctx,
		"INSERT INTO users (name, password, is_superuser, created_at, updated_at) VALUES (?, ?, ?, ?, ?);",
		user.Name,
		user.Password,
		user.IsSuperuser,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return user, db.RollbackError(tx, errors.WithStack(err))
	}
	id, err := res.LastInsertId()
	if err != nil {
		return user, db.RollbackError(tx, errors.WithStack(err))
	}
	user.ID = int(id)
	err = s.auditor.LogCreate(ctx, tx, user)
	if err != nil {
		return user, db.RollbackError(tx, errors.WithStack(err))
	}
	return user, errors.WithStack(tx.Commit())
}

// Update saves an updated entity to the database.
func (s Store) Update(ctx context.Context, user *entity.User) (*entity.User, error) {
	if user.ID < 1 {
		return user, errors.WithStack(db.ErrNotExists)
	}
	current, err := s.Find(ctx, user.ID)
	if err != nil {
		return user, errors.WithStack(err)
	}

	if user.Password == "" {
		user.Password = current.Password
	}
	user.CreatedAt = current.CreatedAt
	user.UpdatedAt = null.TimeFrom(s.clock.Now())

	tx, err := s.db.Begin()
	if err != nil {
		return user, errors.WithStack(err)
	}
	_, err = tx.ExecContext(ctx,
		"UPDATE users SET name = ?, password = ?, is_superuser = ?, updated_at = ? WHERE id = ?",
		user.Name,
		user.Password,
		user.IsSuperuser,
		user.UpdatedAt,
		user.ID,
	)
	if err != nil {
		return user, db.RollbackError(tx, errors.WithStack(err))
	}

	err = s.auditor.LogUpdate(ctx, tx, current, user)
	if err != nil {
		return user, db.RollbackError(tx, errors.WithStack(err))
	}

	return user, errors.WithStack(tx.Commit())
}

// Delete removes multiple entities from the database.
func (s Store) Delete(ctx context.Context, ids []int) ([]*entity.User, error) {
	var result []*entity.User
	sources, err := s.GetByID(ctx, ids)
	if err != nil {
		return result, errors.WithStack(err)
	}
	if len(sources) < 1 {
		return result, nil
	}
	tx, err := s.db.Begin()
	if err != nil {
		return result, errors.WithStack(err)
	}
	for _, source := range sources {
		if r, err := s.deleteEntity(ctx, tx, source); err == nil {
			result = append(result, r)
		} else {
			return result, db.RollbackError(tx, errors.WithStack(err))
		}
	}
	return result, errors.WithStack(tx.Commit())
}

// deleteEntity removes a single entity from the database.
func (s Store) deleteEntity(ctx context.Context, tx *db.Tx, source *entity.User) (*entity.User, error) {
	if source.ID <= 1 {
		return source, errors.New("cannot delete admin user")
	}
	_, err := tx.ExecContext(ctx, "DELETE FROM users WHERE id = ? LIMIT 1", source.ID)
	if err != nil {
		return source, errors.WithStack(err)
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM role_user WHERE user_id = ?", source.ID)
	if err != nil {
		return source, errors.WithStack(err)
	}
	err = s.auditor.LogDelete(ctx, tx, source)
	if err != nil {
		return source, errors.WithStack(err)
	}
	return source, nil
}

// SyncRoles sets the provided role IDs for a user.
// nolint:govet
func (s Store) SyncRoles(ctx context.Context, source *entity.User, roleIDs []int) (*entity.User, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return source, errors.WithStack(err)
	}
	current, err := s.getCurrentRoles(ctx, tx, source)
	if err != nil {
		return source, db.RollbackError(tx, errors.WithStack(err))
	}
	for _, roleID := range current {
		s.auth.RemoveRoleForUser(source.ID, roleID)
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM role_user WHERE user_id = ?", source.ID)
	if err != nil {
		return source, db.RollbackError(tx, errors.WithStack(err))
	}
	for _, roleID := range roleIDs {
		query, params, err := sq.Insert("role_user").SetMap(db.ColumnMap{"user_id": source.ID, "role_id": roleID}).ToSql()
		if err != nil {
			return source, db.RollbackError(tx, errors.WithStack(err))
		}
		_, err = tx.ExecContext(ctx, query, params...)
		if err != nil {
			return source, db.RollbackError(tx, errors.WithStack(err))
		}
		s.auth.AddRoleForUser(source.ID, roleID)
	}
	err = s.auditor.LogSync(ctx, tx, source, "roles", roleIDs, current)
	if err != nil {
		return source, db.RollbackError(tx, errors.WithStack(err))
	}
	return source, errors.WithStack(tx.Commit())
}

// getCurrentRoles returns all currently attached call type ids.
func (s Store) getCurrentRoles(ctx context.Context, tx *db.Tx, source *entity.User) ([]int, error) {
	type result struct {
		RoleID int `json:"role_id"`
	}
	var r []*result
	err := s.db.SelectContext(ctx, &r, "SELECT role_id FROM role_user WHERE user_id = ?", source.ID)
	if err != nil {
		return nil, db.RollbackError(tx, errors.WithStack(err))
	}
	var current []int
	for _, res := range r {
		current = append(current, res.RoleID)
	}
	return current, nil
}

// checkNotFound returns a ErrNotFound if no rows were returned.
func checkNotFound(err error) error {
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	return err
}
