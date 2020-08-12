package entity

import "gopkg.in/guregu/null.v3"

// Role belongs to a user, has many permissions.
type Role struct {
	ID   int    `json:"id"`
	Name string `json:"name"`

	CreatedAt null.Time `json:"created_at" diff:"-"`
	UpdatedAt null.Time `json:"updated_at" diff:"-"`
}

// Primary returns the primary key of this entity.
func (r Role) Primary() int {
	return r.ID
}

// Type returns a string representation of this entity's type.
func (r Role) Type() Kind {
	return KindRole
}
