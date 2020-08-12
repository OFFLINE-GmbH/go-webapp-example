package audit

import (
	"context"

	"go-webapp-example/internal/pkg/entity"
	"go-webapp-example/pkg/db"

	"gopkg.in/guregu/null.v3"
)

var _ ChangeAuditor = &MockAuditor{}

type SystemAuditor interface {
	LogSystem(ctx context.Context, tx *db.Tx, action string, e entity.Entity) error
}

type MockAuditor struct {
	Created []entity.AuditLog
	Updated []entity.AuditLog
	Deleted []entity.AuditLog
	Synced  []entity.AuditLog
}

func NewMockAuditor() *MockAuditor {
	a := &MockAuditor{}
	a.Clear()
	return a
}

func (a *MockAuditor) LogCreate(ctx context.Context, tx *db.Tx, e entity.Entity) error {
	a.Created = append(a.Created, getMockLog(e))
	return nil
}

func (a *MockAuditor) LogUpdate(ctx context.Context, tx *db.Tx, from, e entity.Entity) error {
	a.Updated = append(a.Updated, getMockLog(e))
	return nil
}

func (a *MockAuditor) LogDelete(ctx context.Context, tx *db.Tx, e entity.Entity) error {
	a.Deleted = append(a.Deleted, getMockLog(e))
	return nil
}

func (a *MockAuditor) LogSync(ctx context.Context, tx *db.Tx, e entity.Entity, relation string, valuesNew, valuesOld interface{}) error {
	a.Synced = append(a.Synced, getMockLog(e))
	return nil
}

func (a *MockAuditor) Clear() {
	a.Created = []entity.AuditLog{}
	a.Updated = []entity.AuditLog{}
	a.Deleted = []entity.AuditLog{}
}

func getMockLog(e entity.Entity) entity.AuditLog {
	return entity.AuditLog{
		EntityID:   null.IntFrom(int64(e.Primary())),
		EntityType: e.Type(),
	}
}
