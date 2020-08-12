package audit

import (
	"context"
	"fmt"
	"strings"

	"go-webapp-example/internal/pkg/entity"
	"go-webapp-example/pkg/db"
	"go-webapp-example/pkg/log"
	"go-webapp-example/pkg/session"

	"github.com/pkg/errors"
	"github.com/r3labs/diff"
	"gopkg.in/guregu/null.v3"
)

type ChangeAuditor interface {
	LogCreate(ctx context.Context, tx *db.Tx, e entity.Entity) error
	LogUpdate(ctx context.Context, tx *db.Tx, from, to entity.Entity) error
	LogDelete(ctx context.Context, tx *db.Tx, e entity.Entity) error
	LogSync(ctx context.Context, tx *db.Tx, e entity.Entity, relation string, valuesNew interface{}, valuesOld interface{}) error
}

var _ ChangeAuditor = &Service{}

const (
	ActionCreated = "created"
	ActionUpdated = "updated"
	ActionDeleted = "deleted"

	ActionLoggedIn = "loggedin"
	ActionUnknown  = "unknown"
)

// Service is used to interact with the entity. It
// allows access to the store by embedding it.
type Service struct {
	*Store
	logger log.Logger
}

// NewService returns a pointer to a new Service.
func NewService(store *Store, logger log.Logger) *Service {
	return &Service{
		Store:  store,
		logger: logger,
	}
}

// LogSystem logs a system level event.
func (s Service) LogSystem(ctx context.Context, tx *db.Tx, action string, e entity.Entity) error {
	l := &entity.AuditLog{
		EntityID:   null.IntFrom(int64(e.Primary())),
		EntityType: e.Type(),
		Action:     action,
	}
	return s.persist(ctx, tx, l)
}

// LogCreate creates a log entry for a created entity.
func (s Service) LogCreate(ctx context.Context, tx *db.Tx, e entity.Entity) error {
	l := &entity.AuditLog{
		EntityID:   null.IntFrom(int64(e.Primary())),
		EntityType: e.Type(),
		Action:     ActionCreated,
	}
	return s.persist(ctx, tx, l)
}

// LogUpdate creates a log entry for a changed entity.
func (s Service) LogUpdate(ctx context.Context, tx *db.Tx, from, to entity.Entity) error {
	changelog, _ := diff.Diff(from, to)

	for _, change := range changelog {
		var field string
		if len(change.Path) == 1 {
			field = change.Path[0]
		} else {
			field = fmt.Sprintf("%v", change.Path)
		}

		action := mapDiffAction(change.Type)

		l := &entity.AuditLog{
			EntityID:   null.IntFrom(int64(from.Primary())),
			EntityType: from.Type(),
			Action:     action,
			Field:      strings.ToLower(field),
			ValueNew:   fmt.Sprintf("%v", change.To),
			ValueOld:   fmt.Sprintf("%v", change.From),
		}
		if err := s.persist(ctx, tx, l); err != nil {
			return err
		}
	}
	return nil
}

// LogDelete creates a log entry for a deleted entity.
func (s Service) LogDelete(ctx context.Context, tx *db.Tx, e entity.Entity) error {
	l := &entity.AuditLog{
		EntityID:   null.IntFrom(int64(e.Primary())),
		EntityType: e.Type(),
		Action:     ActionDeleted,
	}
	return s.persist(ctx, tx, l)
}

// LogSync creates a log entry for a synced relationship.
func (s Service) LogSync(ctx context.Context, tx *db.Tx, e entity.Entity, relation string, valuesNew, valuesOld interface{}) error {
	l := &entity.AuditLog{
		EntityID:   null.IntFrom(int64(e.Primary())),
		EntityType: e.Type(),
		Field:      relation,
		ValueOld:   fmt.Sprintf("%v", valuesOld),
		ValueNew:   fmt.Sprintf("%v", valuesNew),
		Action:     ActionUpdated,
	}
	return s.persist(ctx, tx, l)
}

// createLogEntry creates the log entry in the database.
func (s Service) persist(ctx context.Context, tx *db.Tx, l *entity.AuditLog) error {
	// Try to fetch the acting user from the context.
	u, err := session.UserFromContext(ctx)
	if err == nil {
		l.UserID = u.Primary()
	}

	_, err = s.Create(ctx, tx, l)
	if err != nil {
		return errors.Wrapf(err, "failed to create audit log entry: %+v", l)
	}
	return nil
}

// map diff actions to internal action names.
func mapDiffAction(in string) string {
	m := map[string]string{
		diff.CREATE: ActionCreated,
		diff.UPDATE: ActionUpdated,
		diff.DELETE: ActionDeleted,
	}
	action, ok := m[in]
	if !ok {
		return ActionUnknown
	}
	return action
}
