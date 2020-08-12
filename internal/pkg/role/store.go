package role

import (
	"context"
	"database/sql"

	"go-webapp-example/internal/pkg/entity"
	"go-webapp-example/pkg/clock"
	"go-webapp-example/pkg/db"
	"go-webapp-example/pkg/util"

	"github.com/pkg/errors"
	"gopkg.in/guregu/null.v3"

	sq "github.com/Masterminds/squirrel"
)

type authManager interface {
	DeleteRole(roleID int)
	AddRolePermission(roleID int, code string, s string) bool

	AddRoleForUser(userID int, roleID int) bool
	RemoveRoleForUser(userID int, roleID int) bool
}

// Store handles the direct database access for this entity.
type Store struct {
	db    *db.Connection
	clock *clock.Clock
	auth  authManager
}

// NewStore returns a new store instance.
func NewStore(conn *db.Connection, auth authManager, opts ...func(s *Store)) *Store {
	s := &Store{db: conn, auth: auth}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Find finds the entity by id.
func (s Store) Find(ctx context.Context, id int) (*entity.Role, error) {
	var role entity.Role
	err := s.db.GetContext(ctx, &role, "SELECT * FROM roles WHERE id = ? LIMIT 1", id)
	return &role, errors.WithStack(checkNotFound(err))
}

// Get returns all available entities.
func (s Store) Get(ctx context.Context) ([]*entity.Role, error) {
	var roles []*entity.Role
	err := s.db.SelectContext(ctx, &roles, "SELECT * FROM roles")
	return roles, errors.WithStack(err)
}

// GetByID returns entities by ID.
func (s Store) GetByID(ctx context.Context, ids []int) (map[int]*entity.Role, error) {
	var indicators []*entity.Role
	query, params, err := sq.Select("*").From("roles").Where(sq.Eq{"id": util.UniqueInts(ids)}).ToSql()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	err = s.db.SelectContext(ctx, &indicators, query, params...)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	result := make(map[int]*entity.Role, len(indicators))
	for _, item := range indicators {
		result[item.ID] = item
	}
	return result, errors.WithStack(err)
}

// GetByUserID returns a map of user ids to a slice of roles.
func (s Store) GetByUserID(ctx context.Context, ids []int) (map[int][]*entity.Role, error) {
	type result struct {
		entity.Role
		UserID int `json:"user_id"`
	}
	var roles []*result
	ret := make(map[int][]*entity.Role)
	query, params, err := sq.
		Select("roles.*, role_user.user_id as user_id").
		From("role_user").
		LeftJoin("roles ON role_user.role_id = roles.id").
		Where(sq.Eq{"role_user.user_id": ids}).
		ToSql()
	if err != nil {
		return ret, errors.WithStack(err)
	}
	err = s.db.SelectContext(ctx, &roles, query, params...)
	if err != nil {
		return ret, errors.WithStack(err)
	}
	for _, role := range roles {
		ret[role.UserID] = append(ret[role.UserID], &entity.Role{
			ID:        role.ID,
			Name:      role.Name,
			CreatedAt: role.CreatedAt,
			UpdatedAt: role.UpdatedAt,
		})
	}
	return ret, errors.WithStack(checkNotFound(err))
}

// Create creates a new entity.
func (s Store) Create(ctx context.Context, role *entity.Role) (*entity.Role, error) {
	role.CreatedAt = null.TimeFrom(s.clock.Now())
	role.UpdatedAt = null.TimeFrom(s.clock.Now())
	res, err := s.db.Exec("INSERT INTO roles (name, created_at, updated_at) VALUES (?, ?, ?);", role.Name, role.CreatedAt, role.UpdatedAt)
	if err != nil {
		return role, errors.WithStack(err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return role, errors.WithStack(err)
	}
	role.ID = int(id)
	return role, nil
}

// Update saves an updated entity to the database.
func (s Store) Update(ctx context.Context, role *entity.Role) (*entity.Role, error) {
	if role.ID < 1 {
		return role, errors.WithStack(db.ErrNotExists)
	}
	role.UpdatedAt = null.TimeFrom(s.clock.Now())
	_, err := s.db.Exec("UPDATE roles SET name = ?, updated_at = ? WHERE id = ?", role.Name, role.UpdatedAt, role.ID)
	return role, errors.WithStack(err)
}

// Delete removes multiple entities from the database.
func (s Store) Delete(ctx context.Context, ids []int) ([]*entity.Role, error) {
	var roles []*entity.Role
	if len(ids) < 1 {
		return roles, nil
	}
	for _, id := range ids {
		if id == 1 {
			return roles, errors.New("cannot delete admin role")
		}
	}
	returned, err := s.GetByID(ctx, ids)
	if err != nil {
		return roles, errors.WithStack(err)
	}
	tx, err := s.db.Begin()
	if err != nil {
		return roles, errors.WithStack(err)
	}
	query, params, err := sq.Delete("roles").Where(sq.Eq{"id": util.UniqueInts(ids)}).ToSql()
	if err != nil {
		return nil, db.RollbackError(tx, errors.WithStack(err))
	}
	_, err = tx.ExecContext(ctx, query, params...)
	if err != nil {
		return nil, db.RollbackError(tx, errors.WithStack(err))
	}
	query, params, err = sq.Delete("role_user").Where(sq.Eq{"role_id": util.UniqueInts(ids)}).ToSql()
	if err != nil {
		return nil, db.RollbackError(tx, errors.WithStack(err))
	}
	_, err = tx.ExecContext(ctx, query, params...)
	if err != nil {
		return nil, db.RollbackError(tx, errors.WithStack(err))
	}
	for _, r := range returned {
		roles = append(roles, r)
	}
	for _, id := range ids {
		s.auth.DeleteRole(id)
	}
	return roles, tx.Commit()
}

// SyncUsers sets the provided user IDs for a role.
// nolint:govet
func (s Store) SyncUsers(ctx context.Context, source *entity.Role, userIDs []int) (*entity.Role, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return source, errors.WithStack(err)
	}
	current, err := s.getCurrentUsers(ctx, tx, source)
	if err != nil {
		return source, db.RollbackError(tx, errors.WithStack(err))
	}
	for _, userID := range current {
		s.auth.RemoveRoleForUser(userID, source.ID)
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM role_user WHERE role_id = ?", source.ID)
	if err != nil {
		return source, db.RollbackError(tx, errors.WithStack(err))
	}
	for _, userID := range userIDs {
		query, params, err := sq.Insert("role_user").SetMap(db.ColumnMap{"user_id": userID, "role_id": source.ID}).ToSql()
		if err != nil {
			return source, db.RollbackError(tx, errors.WithStack(err))
		}
		_, err = tx.ExecContext(ctx, query, params...)
		if err != nil {
			return source, db.RollbackError(tx, errors.WithStack(err))
		}
		s.auth.AddRoleForUser(userID, source.ID)
	}
	return source, errors.WithStack(tx.Commit())
}

// getCurrentUsers returns all currently attached call type ids.
func (s Store) getCurrentUsers(ctx context.Context, tx *db.Tx, source *entity.Role) ([]int, error) {
	type result struct {
		UserID int `json:"user_id"`
	}
	var r []*result
	err := s.db.SelectContext(ctx, &r, "SELECT user_id FROM role_user WHERE role_id = ?", source.ID)
	if err != nil {
		return nil, db.RollbackError(tx, errors.WithStack(err))
	}
	var current []int
	for _, res := range r {
		current = append(current, res.UserID)
	}
	return current, nil
}

// permissionsMap contains all permissions with a map of their active levels.
type permissionsMap map[string]map[entity.PermissionLevel]bool

// SyncPermissions sets the permissions for a role. It makes sure that all lower levels are also present for easier assertions.
func (s Store) SyncPermissions(ctx context.Context, u *entity.Role, permissions []*entity.Permission) (*entity.Role, error) {
	s.auth.DeleteRole(u.ID)
	perms := make(permissionsMap)
	// Make sure all lower permissions are included as well.
	for _, permission := range permissions {
		if permission.Level == entity.PermissionLevelNone {
			continue
		}
		if _, ok := perms[permission.Code]; !ok {
			perms[permission.Code] = make(map[entity.PermissionLevel]bool)
		}
		perms = ensureLevelPermissions(perms, permission)
	}
	// Add the permissions for the role.
	for code, levels := range perms {
		for level := range levels {
			s.auth.AddRolePermission(u.ID, code, string(level))
		}
	}
	return u, nil
}

// ensureLevelPermissions make sure that all lower levels for a permission are included as well.
func ensureLevelPermissions(perms permissionsMap, permission *entity.Permission) permissionsMap {
	levels := []entity.PermissionLevel{entity.PermissionLevelRead, entity.PermissionLevelWrite, entity.PermissionLevelManage}
	if permission.Level == entity.PermissionLevelWrite {
		levels = levels[:2]
	} else if permission.Level == entity.PermissionLevelRead {
		levels = levels[:1]
	}
	for _, level := range levels {
		perms[permission.Code][level] = true
	}
	return perms
}

// checkNotFound returns a ErrNotFound if no rows were returned.
func checkNotFound(err error) error {
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	return err
}
