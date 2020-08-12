package permission

import (
	"github.com/pkg/errors"
)

// ErrNotFound is returned when a requested permission could not be found.
var ErrNotFound = errors.New("permission not found")
