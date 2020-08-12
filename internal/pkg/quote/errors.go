package quote

import (
	"github.com/pkg/errors"
)

// ErrNotFound is returned when a requested quote could not be found.
var ErrNotFound = errors.New("quote not found")
