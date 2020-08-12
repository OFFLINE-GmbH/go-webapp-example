package permission

import (
	"context"
	"database/sql"

	"go-webapp-example/internal/pkg/entity"
	"go-webapp-example/pkg/clock"
	"go-webapp-example/pkg/db"

	"github.com/pkg/errors"
)

type authManager interface {
	PermissionsForUser(id int) [][]string
	PermissionsForRole(id int) [][]string
}

// Store handles the direct database access for this entity.
type Store struct {
	db    *db.Connection
	auth  authManager
	clock *clock.Clock
}

// NewStore returns a new store instance.
func NewStore(conn *db.Connection, authManager authManager, opts ...func(s *Store)) *Store {
	s := &Store{db: conn, auth: authManager}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Find finds the entity by id.
func (s Store) Find(ctx context.Context, id int) (*entity.Permission, error) {
	var permission entity.Permission
	err := s.db.GetContext(ctx, &permission, "SELECT * FROM permissions WHERE id = ? LIMIT 1", id)
	return &permission, errors.WithStack(checkNotFound(err))
}

// Get returns all available entities.
func (s Store) Get(ctx context.Context) ([]*entity.Permission, error) {
	var permissions []*entity.Permission
	err := s.db.SelectContext(ctx, &permissions, "SELECT * FROM permissions")
	return permissions, errors.WithStack(err)
}

// GetByUserID returns a map of user ids to a slice of permissions.
func (s Store) GetByUserID(ctx context.Context, ids []int) (map[int][]*entity.Permission, error) {
	ret := make(map[int][]*entity.Permission)
	for _, id := range ids {
		perms := s.auth.PermissionsForUser(id)
		for _, perm := range perms {
			ret[id] = append(ret[id], &entity.Permission{
				Code:  perm[1],
				Level: entity.PermissionLevel(perm[2]),
			})
		}
	}
	return ret, nil
}

// GetForUserID returns a slice of permissions for a given user.
func (s Store) GetForUserID(ctx context.Context, id int) []*entity.Permission {
	var ret []*entity.Permission
	perms := s.auth.PermissionsForUser(id)
	for _, perm := range perms {
		ret = append(ret, &entity.Permission{
			Code:  perm[1],
			Level: entity.PermissionLevel(perm[2]),
		})
	}
	return ret
}

// GetByRoleID returns a map of role ids to a slice of permissions.
func (s Store) GetByRoleID(ctx context.Context, ids []int) (map[int][]*entity.Permission, error) {
	ret := make(map[int][]*entity.Permission)
	for _, id := range ids {
		perms := s.auth.PermissionsForRole(id)
		for _, perm := range perms {
			ret[id] = append(ret[id], &entity.Permission{
				Code:  perm[1],
				Level: entity.PermissionLevel(perm[2]),
			})
		}
	}
	return ret, nil
}

// Create creates a new entity.
func (s Store) Create(ctx context.Context, permission *entity.Permission) (*entity.Permission, error) {
	res, err := s.db.Exec("INSERT INTO permissions (code) VALUES (?);", permission.Code)
	if err != nil {
		return permission, errors.WithStack(err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return permission, errors.WithStack(err)
	}
	permission.ID = int(id)
	return permission, nil
}

// Update saves an updated entity to the database.
func (s Store) Update(ctx context.Context, permission *entity.Permission) (*entity.Permission, error) {
	if permission.ID < 1 {
		return permission, errors.WithStack(db.ErrNotExists)
	}
	_, err := s.db.Exec("UPDATE permissions SET code = ? WHERE id = ?", permission.Code, permission.ID)
	return permission, errors.WithStack(err)
}

// Delete removes an entity from the database.
func (s Store) Delete(ctx context.Context, permission *entity.Permission) (*entity.Permission, error) {
	if permission.ID < 1 {
		return permission, nil
	}
	_, err := s.db.Exec("DELETE FROM permissions WHERE id = ? LIMIT 1", permission.ID)
	return permission, errors.WithStack(err)
}

// checkNotFound returns a ErrNotFound if no rows were returned.
func checkNotFound(err error) error {
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	return err
}
