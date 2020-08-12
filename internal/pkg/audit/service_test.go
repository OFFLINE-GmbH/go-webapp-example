package audit

import (
	"context"
	"testing"
	"time"

	"go-webapp-example/internal/pkg/entity"
	"go-webapp-example/internal/pkg/test"
	"go-webapp-example/pkg/clock"
	"go-webapp-example/pkg/db"
	"go-webapp-example/pkg/log"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v3"
)

// cols are the default DB columns.
var cols = []string{
	"id",
	"user_id",
	"value_old",
	"value_new",
	"action",
	"entity_type",
	"entity_id",
	"meta",
	"created_at",
	"updated_at",
}

// now is used as time for all test cases.
var now = time.Now()

// TestAuditService tests all service methods as well as the underlying store.
func TestAuditService(t *testing.T) {
	dbConn, mock := test.MockDB(t)
	mock.ExpectBegin()
	tx := &db.Tx{Tx: dbConn.MustBegin(), Log: log.NewNullLogger()}

	service := NewService(NewStore(dbConn, func(s *Store) {
		s.clock = clock.FromTime(now)
	}), log.NewNullLogger())

	t.Run("Get", get(mock, service))
	t.Run("Find", find(mock, service))
	t.Run("Create", create(mock, tx, service))
	t.Run("Update", update(mock, tx, service))
	t.Run("Delete", del(mock, tx, service))
	t.Run("LogSystem", logSystem(mock, tx, service))
	t.Run("LogCreate", logCreate(mock, tx, service))
	t.Run("LogUpdate", logChange(mock, tx, service))
	t.Run("LogDelete", logDelete(mock, tx, service))
	t.Run("LogSync", logSync(mock, tx, service))
}

func get(mock sqlmock.Sqlmock, service *Service) func(t *testing.T) {
	return func(t *testing.T) {
		rows := sqlmock.
			NewRows(cols).
			AddRow(1, 1, "old value", "new value", "updated", "log", 5, "", time.Now(), time.Now()).
			AddRow(2, 1, "", "", "created", "log", 5, "", time.Now(), nil)

		mock.ExpectQuery("SELECT .+ FROM auditlogs").WillReturnRows(rows)

		auditlogs, err := service.Get(context.Background())

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.Len(t, auditlogs, 2)
	}
}

func find(mock sqlmock.Sqlmock, service *Service) func(t *testing.T) {
	return func(t *testing.T) {
		rows := sqlmock.
			NewRows(cols).
			AddRow(1, 1, "old value", "new value", "updated", "log", 5, "", time.Now(), time.Now())

		mock.
			ExpectQuery("SELECT .+ FROM auditlogs WHERE id = . LIMIT 1").
			WithArgs(1).
			WillReturnRows(rows)

		l, err := service.Find(context.Background(), 1)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.Equal(t, l.ID, 1)
	}
}

func create(mock sqlmock.Sqlmock, tx *db.Tx, service *Service) func(t *testing.T) {
	return func(t *testing.T) {
		l := &entity.AuditLog{
			Action:   ActionUpdated,
			EntityID: null.IntFrom(1),
			Meta:     "meta",
			UserID:   1,
			ValueOld: "old",
			ValueNew: "new",
			Field:    "name",
		}

		mock.
			ExpectExec("INSERT INTO auditlogs").
			WithArgs(
				l.Action,
				now,
				l.EntityID,
				l.EntityType,
				l.Field,
				l.Meta,
				now,
				l.UserID,
				l.ValueNew,
				l.ValueOld,
			).
			WillReturnResult(sqlmock.NewResult(3, 1))

		result, err := service.Create(context.Background(), tx, l)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.False(t, l.CreatedAt.IsZero())
		assert.False(t, l.UpdatedAt.IsZero())
		assert.NotEqual(t, 0, result.ID)
	}
}

func update(mock sqlmock.Sqlmock, tx *db.Tx, service *Service) func(t *testing.T) {
	return func(t *testing.T) {
		l := &entity.AuditLog{
			ID:        3,
			Action:    ActionUpdated,
			CreatedAt: null.TimeFrom(now),
			EntityID:  null.IntFrom(1),
			Meta:      "meta",
			UserID:    1,
			ValueNew:  "new",
			ValueOld:  "old",
			Field:     "name",
		}

		mock.
			ExpectExec(`UPDATE auditlogs .+ WHERE id = .`).
			WithArgs(
				l.Action,
				now,
				l.EntityID,
				l.EntityType,
				l.Field,
				l.Meta,
				now,
				l.UserID,
				l.ValueNew,
				l.ValueOld,
				l.ID,
			).
			WillReturnResult(sqlmock.NewResult(0, 1))

		_, err := service.Update(context.Background(), tx, l)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.False(t, l.UpdatedAt.IsZero())
	}
}

func del(mock sqlmock.Sqlmock, tx *db.Tx, service *Service) func(t *testing.T) {
	return func(t *testing.T) {
		mock.
			ExpectExec("DELETE FROM auditlogs WHERE id = . LIMIT 1").
			WithArgs(3).
			WillReturnResult(sqlmock.NewResult(0, 1))

		l := &entity.AuditLog{ID: 3}

		_, err := service.Delete(context.Background(), tx, l)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	}
}

func logSystem(mock sqlmock.Sqlmock, tx *db.Tx, service SystemAuditor) func(t *testing.T) {
	return func(t *testing.T) {
		mock.
			ExpectExec("INSERT INTO auditlogs").
			WithArgs(
				ActionLoggedIn,
				now,
				1,
				entity.KindUser,
				"",
				"",
				now,
				0,
				"",
				"",
			).
			WillReturnResult(sqlmock.NewResult(3, 1))

		err := service.LogSystem(context.Background(), tx, ActionLoggedIn, entity.User{ID: 1})

		assert.NoError(t, mock.ExpectationsWereMet())
		assert.NoError(t, err)
	}
}

func logCreate(mock sqlmock.Sqlmock, tx *db.Tx, service *Service) func(t *testing.T) {
	return func(t *testing.T) {
		mock.
			ExpectExec("INSERT INTO auditlogs").
			WithArgs(
				ActionCreated,
				now,
				2,
				entity.KindUser,
				"",
				"",
				now,
				0,
				"",
				"",
			).
			WillReturnResult(sqlmock.NewResult(3, 1))

		from := entity.User{
			ID:   2,
			Name: "New",
		}

		err := service.LogCreate(context.Background(), tx, from)

		assert.NoError(t, mock.ExpectationsWereMet())
		assert.NoError(t, err)
	}
}

func logChange(mock sqlmock.Sqlmock, tx *db.Tx, service *Service) func(t *testing.T) {
	return func(t *testing.T) {
		mock.
			ExpectExec("INSERT INTO auditlogs").
			WithArgs(
				ActionUpdated,
				now,
				2,
				entity.KindQuote,
				"author",
				"",
				now,
				0,
				"New Name",
				"Old Name",
			).
			WillReturnResult(sqlmock.NewResult(3, 1))

		from := entity.Quote{ID: 2, Author: "Old Name"}
		to := entity.Quote{ID: 2, Author: "New Name"}

		err := service.LogUpdate(context.Background(), tx, from, to)

		assert.NoError(t, mock.ExpectationsWereMet())
		assert.NoError(t, err)
	}
}

func logDelete(mock sqlmock.Sqlmock, tx *db.Tx, service *Service) func(t *testing.T) {
	return func(t *testing.T) {
		mock.
			ExpectExec("INSERT INTO auditlogs").
			WithArgs(
				ActionDeleted,
				now,
				2,
				entity.KindUser,
				"",
				"",
				now,
				0,
				"",
				"",
			).
			WillReturnResult(sqlmock.NewResult(3, 1))

		from := entity.User{ID: 2, Name: "Deleted"}

		err := service.LogDelete(context.Background(), tx, from)

		assert.NoError(t, mock.ExpectationsWereMet())
		assert.NoError(t, err)
	}
}

func logSync(mock sqlmock.Sqlmock, tx *db.Tx, service *Service) func(t *testing.T) {
	return func(t *testing.T) {
		mock.
			ExpectExec("INSERT INTO auditlogs").
			WithArgs(
				ActionUpdated,
				now,
				2,
				entity.KindQuote,
				"roles",
				"",
				now,
				0,
				"[1 2 3]",
				"[4]",
			).
			WillReturnResult(sqlmock.NewResult(3, 1))

		to := entity.Quote{
			ID:     2,
			Author: "New Name",
		}

		err := service.LogSync(context.Background(), tx, to, "roles", []int{1, 2, 3}, []int{4})

		assert.NoError(t, mock.ExpectationsWereMet())
		assert.NoError(t, err)
	}
}
