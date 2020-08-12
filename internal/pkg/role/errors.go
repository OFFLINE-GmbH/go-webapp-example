package role

import (
	"github.com/pkg/errors"
)

// ErrNotFound is returned when a requested role could not be found.
var ErrNotFound = errors.New("role not found")
