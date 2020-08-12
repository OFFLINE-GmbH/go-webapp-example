package role

import (
	"context"
	"testing"
	"time"

	"go-webapp-example/internal/pkg/entity"
	"go-webapp-example/internal/pkg/test"
	"go-webapp-example/pkg/clock"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// now is used as time for all test cases.
var now = time.Now()

type authManagerMock struct{}

func (a authManagerMock) AddRolePermission(id int, code, s string) bool { return true }
func (a authManagerMock) AddRoleForUser(userID, roleID int) bool        { return true }
func (a authManagerMock) RemoveRoleForUser(userID, roleID int) bool     { return true }
func (a authManagerMock) DeleteRole(id int)                             {}

// TestRoleService tests all service methods as well as the underlying store.
func TestRoleService(t *testing.T) {
	db, mock := test.MockDB(t)

	authMock := &authManagerMock{}
	service := NewService(NewStore(db, authMock, func(store *Store) {
		store.clock = clock.FromTime(now)
	}))

	t.Run("Get", get(mock, service))
	t.Run("GetByID", getByID(mock, service))
	t.Run("Create", create(mock, service))
	t.Run("Update", update(mock, service))
	t.Run("Delete", del(mock, service))
	t.Run("GetByUserID", getByUserID(mock, service))
}

func get(mock sqlmock.Sqlmock, service *Service) func(t *testing.T) {
	return func(t *testing.T) {
		rows := sqlmock.
			NewRows([]string{"id", "name", "created_at", "updated_at"}).
			AddRow(1, "admin", time.Now(), time.Now()).
			AddRow(2, "role", time.Now(), time.Now())

		mock.ExpectQuery("SELECT (.+) FROM roles$").WillReturnRows(rows)

		roles, err := service.Get(context.Background())

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.Len(t, roles, 2)
	}
}

func getByID(mock sqlmock.Sqlmock, service *Service) func(t *testing.T) {
	return func(t *testing.T) {
		rows := sqlmock.
			NewRows([]string{"id", "name", "created_at", "updated_at"}).
			AddRow(1, "admin", time.Now(), time.Now()).
			AddRow(2, "role", time.Now(), time.Now())

		mock.ExpectQuery("SELECT .+ FROM roles WHERE id IN \\(.*\\)").
			WithArgs(1, 2).
			WillReturnRows(rows)

		indicators, err := service.GetByID(context.Background(), []int{1, 2})

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.Len(t, indicators, 2)
	}
}

func create(mock sqlmock.Sqlmock, service *Service) func(t *testing.T) {
	return func(t *testing.T) {
		mock.
			ExpectExec("INSERT INTO roles").
			WithArgs("Created", now, now).
			WillReturnResult(sqlmock.NewResult(3, 1))

		role := &entity.Role{Name: "Created"}

		result, err := service.Create(context.Background(), role)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.NotEqual(t, 0, result.ID)
		assert.False(t, role.CreatedAt.IsZero())
		assert.False(t, role.UpdatedAt.IsZero())
	}
}

func update(mock sqlmock.Sqlmock, service *Service) func(t *testing.T) {
	return func(t *testing.T) {
		mock.
			ExpectExec("UPDATE roles SET name = .+, updated_at = .+ WHERE id = .").
			WithArgs("New Role", now, 3).
			WillReturnResult(sqlmock.NewResult(0, 1))

		role := &entity.Role{ID: 3, Name: "New Role"}

		_, err := service.Update(context.Background(), role)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.False(t, role.UpdatedAt.IsZero())
	}
}

func del(mock sqlmock.Sqlmock, service *Service) func(t *testing.T) {
	return func(t *testing.T) {
		mock.
			ExpectQuery("SELECT .+ FROM roles WHERE id IN \\(.*\\)").
			WithArgs(3, 4).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(3, now, now).AddRow(4, now, now))
		mock.ExpectBegin()
		mock.
			ExpectExec("DELETE FROM roles WHERE id IN \\(.*\\)").
			WithArgs(3, 4).
			WillReturnResult(sqlmock.NewResult(0, 2))
		mock.
			ExpectExec("DELETE FROM role_user WHERE role_id IN \\(.*\\)").
			WithArgs(3, 4).
			WillReturnResult(sqlmock.NewResult(0, 2))
		mock.ExpectCommit()
		_, err := service.Delete(context.Background(), []int{3, 4})

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	}
}

func getByUserID(mock sqlmock.Sqlmock, service *Service) func(t *testing.T) {
	return func(t *testing.T) {
		mock.
			ExpectQuery("SELECT .+, role_user.user_id as user_id FROM role_user LEFT JOIN roles ON role_user.role_id = roles.id WHERE role_user.user_id IN .").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "user_id"}).AddRow(1, "role", 4))

		roles, err := service.GetByUserID(context.Background(), []int{1})

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.Equal(t, 1, roles[4][0].ID)
	}
}
