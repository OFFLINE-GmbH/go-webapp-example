package entity

import (
	"github.com/pkg/errors"
	"gopkg.in/guregu/null.v3"
)

var ErrUserInvalidPassword = errors.New("wrong password")

// User is the central user identity used for authentication.
type User struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Password    string `json:"-"`
	IsSuperuser bool   `json:"is_superuser"`

	CreatedAt null.Time `json:"created_at" diff:"-"`
	UpdatedAt null.Time `json:"updated_at" diff:"-"`
}

// Primary returns the primary key of this entity.
func (u User) Primary() int {
	return u.ID
}

// Type returns a string representation of this entity's type.
func (u User) Type() Kind {
	return KindUser
}
