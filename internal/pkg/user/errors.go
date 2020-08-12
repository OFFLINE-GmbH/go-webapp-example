package user

import (
	"github.com/pkg/errors"
)

// ErrNotFound is returned when a requested user could not be found.
var ErrNotFound = errors.New("user not found")
