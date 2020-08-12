package entity

import "fmt"

// Permission is a single action a user can execute. It belongs to
// one or many roles.
type Permission struct {
	ID    int             `json:"id"`
	Code  string          `json:"code"`
	Level PermissionLevel `json:"level"`
}

type PermissionLevel string

const (
	PermissionLevelNone   PermissionLevel = "none"
	PermissionLevelRead   PermissionLevel = "read"
	PermissionLevelWrite  PermissionLevel = "write"
	PermissionLevelManage PermissionLevel = "manage"
)

// Primary returns the primary key of this entity.
func (p Permission) Primary() int {
	return p.ID
}

// Type returns a string representation of this entity's type.
func (p Permission) Type() Kind {
	return KindPermission
}

// CodeLevel returns the combined code and level as a string.
func (p Permission) CodeLevel() string {
	return fmt.Sprintf("%s::%s", p.Code, p.Level)
}
