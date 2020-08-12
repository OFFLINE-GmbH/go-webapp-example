package permission

import (
	"context"
	"testing"
	"time"

	"go-webapp-example/internal/pkg/entity"
	"go-webapp-example/internal/pkg/test"
	"go-webapp-example/pkg/auth"
	"go-webapp-example/pkg/clock"
	"go-webapp-example/pkg/log"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// now is used as time for all test cases.
var now = time.Now()

// TestPermissionService tests all service methods as well as the underlying store.
func TestPermissionService(t *testing.T) {
	db, mock := test.MockDB(t)
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS policies .*").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("SELECT 1 FROM `policies` LIMIT 1").WillReturnResult(sqlmock.NewResult(0, 0))

	authManager, err := auth.New(db, log.NewNullLogger())
	if err != nil {
		t.Fatalf("failed to create auth manager: %s", err)
	}
	service := NewService(NewStore(db, authManager, func(store *Store) {
		store.clock = clock.FromTime(now)
	}))

	t.Run("Get", get(mock, service))
	t.Run("Create", create(mock, service))
	t.Run("Update", update(mock, service))
	t.Run("Delete", del(mock, service))
}

func get(mock sqlmock.Sqlmock, service *Service) func(t *testing.T) {
	return func(t *testing.T) {
		rows := sqlmock.
			NewRows([]string{"id", "code"}).
			AddRow(1, "device.create").
			AddRow(2, "device.update")

		mock.ExpectQuery("SELECT (.+) FROM permissions$").WillReturnRows(rows)

		permissions, err := service.Get(context.Background())

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.Len(t, permissions, 2)
	}
}

func create(mock sqlmock.Sqlmock, service *Service) func(t *testing.T) {
	return func(t *testing.T) {
		mock.
			ExpectExec("INSERT INTO permissions").
			WithArgs("device.delete").
			WillReturnResult(sqlmock.NewResult(3, 1))

		permission := &entity.Permission{Code: "device.delete"}

		result, err := service.Create(context.Background(), permission)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.NotEqual(t, 0, result.ID)
	}
}

func update(mock sqlmock.Sqlmock, service *Service) func(t *testing.T) {
	return func(t *testing.T) {
		mock.
			ExpectExec("UPDATE permissions SET code = .+ WHERE id = .").
			WithArgs("New Permission", 3).
			WillReturnResult(sqlmock.NewResult(0, 1))

		permission := &entity.Permission{ID: 3, Code: "New Permission"}

		_, err := service.Update(context.Background(), permission)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	}
}

func del(mock sqlmock.Sqlmock, service *Service) func(t *testing.T) {
	return func(t *testing.T) {
		mock.
			ExpectExec("DELETE FROM permissions WHERE id = . LIMIT 1").
			WithArgs(3).
			WillReturnResult(sqlmock.NewResult(0, 1))

		permission := &entity.Permission{ID: 3}

		_, err := service.Delete(context.Background(), permission)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	}
}
