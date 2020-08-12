package entity

import "gopkg.in/guregu/null.v3"

// AuditLog represents an audit log entry.
type AuditLog struct {
	ID       int    `json:"id"`
	UserID   int    `json:"user_id"`
	Field    string `json:"field"`
	ValueOld string `json:"value_old"`
	ValueNew string `json:"value_new"`
	// The action describes the kind of action logged (created, updated, deleted).
	Action string `json:"action"`
	// EntityType contains the primary key of the edited entity.
	EntityType Kind `json:"entity_type"`
	// EntityID contains the primary key of the edited entity.
	EntityID null.Int `json:"entity_id"`
	// Meta is used to provide any additional information for this change.
	Meta string `json:"meta"`

	CreatedAt null.Time `json:"created_at"`
	UpdatedAt null.Time `json:"updated_at"`

	User User `json:"user"`
}

// Primary returns the primary key of this entity.
func (l AuditLog) Primary() int {
	return l.ID
}
