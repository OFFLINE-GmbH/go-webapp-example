package user

import (
	"context"
	"testing"
	"time"

	"go-webapp-example/internal/pkg/audit"
	"go-webapp-example/internal/pkg/entity"
	"go-webapp-example/internal/pkg/test"
	"go-webapp-example/pkg/clock"
	"go-webapp-example/pkg/session"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// adminPw is a bcrypt hash of the encrypted password "admin".
const adminPw = "$2y$12$xobp.ipgSFK4keIys2lV3eHB2mV3MUbskFf/OTeXNPObC3xXgrwKm"

// now is used as time for all test cases.
var now = time.Now()

type setupFn func() (sqlmock.Sqlmock, *Service, *audit.MockAuditor)

type authManagerMock struct{}

func (a authManagerMock) AddRoleForUser(int, int) bool    { return true }
func (a authManagerMock) RemoveRoleForUser(int, int) bool { return true }

// TestUserService tests all service methods as well as the underlying store.
func TestUserService(t *testing.T) {
	setup := func() (sqlmock.Sqlmock, *Service, *audit.MockAuditor) {
		db, mockDB := test.MockDB(t)
		mockAuditor := audit.NewMockAuditor()
		authMock := &authManagerMock{}
		service := NewService(
			NewStore(db, authMock, mockAuditor, func(store *Store) {
				store.clock = clock.FromTime(now)
			}),
			session.NewMockSessionManager(),
		)

		return mockDB, service, mockAuditor
	}

	t.Run("Get", get(setup))
	t.Run("GetByID", getByID(setup))
	t.Run("Create", create(setup))
	t.Run("Update", update(setup))
	t.Run("Delete", del(setup))
	t.Run("FindByName", findByName(setup))
	t.Run("Login", login(setup))
	t.Run("ValidateWrongLogin", validateWrongLogin(setup))
}

func get(setup setupFn) func(t *testing.T) {
	return func(t *testing.T) {
		mock, service, _ := setup()

		rows := sqlmock.
			NewRows([]string{"id", "name", "password"}).
			AddRow(1, "admin", "admin").
			AddRow(2, "user", "1234")

		mock.ExpectQuery("SELECT (.+) FROM users$").WillReturnRows(rows)

		users, err := service.Get(context.Background())

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.Len(t, users, 2)
	}
}

func getByID(setup setupFn) func(t *testing.T) {
	return func(t *testing.T) {
		mock, service, _ := setup()
		rows := sqlmock.
			NewRows([]string{"id", "name", "password"}).
			AddRow(1, "admin", "admin").
			AddRow(2, "user", "1234")

		mock.ExpectQuery("SELECT .+ FROM users WHERE id IN \\(.*\\)").
			WithArgs(1, 2).
			WillReturnRows(rows)

		buttons, err := service.GetByID(context.Background(), []int{1, 2})

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.Len(t, buttons, 2)
	}
}

func create(setup setupFn) func(t *testing.T) {
	return func(t *testing.T) {
		mock, service, auditor := setup()

		mock.ExpectBegin()
		mock.
			ExpectExec("INSERT INTO users").
			WithArgs("Edited", "password", true, now, now).
			WillReturnResult(sqlmock.NewResult(3, 1))
		mock.ExpectCommit()

		user := &entity.User{Name: "Edited", Password: "password", IsSuperuser: true}

		// Use the store to circumvent password hashing.
		result, err := service.Store.Create(context.Background(), user)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.NotEqual(t, 0, result.ID)
		assert.Len(t, auditor.Created, 1)
		assert.False(t, user.CreatedAt.IsZero())
		assert.False(t, user.UpdatedAt.IsZero())
	}
}

func update(setup setupFn) func(t *testing.T) {
	return func(t *testing.T) {
		mock, service, auditor := setup()

		mock.
			ExpectQuery("SELECT .+ FROM users WHERE id = .+ LIMIT 1").
			WithArgs(3).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(3, "Old User"))

		mock.ExpectBegin()
		mock.
			ExpectExec("UPDATE users SET name = .+, password = .+, is_superuser = .+, updated_at = .+ WHERE id = ?").
			WithArgs("New User", "password", true, now, 3).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		user := &entity.User{ID: 3, Name: "New User", Password: "password", IsSuperuser: true}

		_, err := service.Store.Update(context.Background(), user)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.Len(t, auditor.Updated, 1)
		assert.False(t, user.UpdatedAt.IsZero())
	}
}

func del(setup setupFn) func(t *testing.T) {
	return func(t *testing.T) {
		mock, service, auditor := setup()
		mock.
			ExpectQuery("SELECT .+ FROM users WHERE id IN \\(.*\\)").
			WithArgs(3, 4).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(3, now, now).AddRow(4, now, now))
		mock.ExpectBegin()
		mock.
			ExpectExec("DELETE FROM users WHERE id = . LIMIT 1").
			WithArgs(sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.
			ExpectExec("DELETE FROM role_user WHERE user_id = .").
			WithArgs(sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.
			ExpectExec("DELETE FROM users WHERE id = . LIMIT 1").
			WithArgs(sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.
			ExpectExec("DELETE FROM role_user WHERE user_id = .").
			WithArgs(sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		_, err := service.Delete(context.Background(), []int{3, 4})

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.Len(t, auditor.Deleted, 2)
	}
}

func findByName(setup setupFn) func(t *testing.T) {
	return func(t *testing.T) {
		mock, service, _ := setup()

		mock.
			ExpectQuery("SELECT .+ FROM users WHERE name = . LIMIT 1").
			WithArgs("admin").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "password"}).AddRow(1, "admin", "password"))

		user, err := service.FindByName(context.Background(), "admin")

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.Equal(t, 1, user.ID)
	}
}

func login(setup setupFn) func(t *testing.T) {
	return func(t *testing.T) {
		mock, service, _ := setup()

		mock.
			ExpectQuery("SELECT .+ FROM users WHERE name = . LIMIT 1").
			WithArgs("admin").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "password"}).AddRow(1, "admin", adminPw))

		user, err := service.Login(context.Background(), "admin", "admin")

		assert.NoError(t, mock.ExpectationsWereMet())
		assert.Equal(t, 1, user.ID)
		assert.NoError(t, err)
	}
}

func validateWrongLogin(setup setupFn) func(t *testing.T) {
	return func(t *testing.T) {
		mock, service, _ := setup()

		mock.
			ExpectQuery("SELECT .+ FROM users WHERE name = . LIMIT 1").
			WithArgs("admin").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "password"}).AddRow(1, "admin", adminPw))

		user, err := service.Login(context.Background(), "admin", "wrong")

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.Nil(t, user, "Login() returned user for wrong password.")
	}
}
