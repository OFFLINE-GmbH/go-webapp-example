package audit

import (
	"github.com/pkg/errors"
)

// ErrNotFound is returned when a requested log could not be found.
var ErrNotFound = errors.New("log not found")
