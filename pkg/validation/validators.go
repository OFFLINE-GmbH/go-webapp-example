package validation

import (
	"fmt"
	"strconv"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
)

// ValidateIP validates a IPv4 input string.
func ValidateIP(input string) bool {
	parts := strings.Split(input, ".")
	if len(parts) != 4 {
		return false
	}
	for _, part := range parts {
		i, err := strconv.Atoi(part)
		if err != nil {
			return false
		}

		if i < 0 || i > 255 {
			return false
		}
	}
	return true
}

// uniqueGetter is used to check for unique DB values.
type uniqueGetter interface {
	Get(target interface{}, query string, args ...interface{}) error
}

// ValidateUniqueInput holds all required values for the ValidateUnique validator.
type ValidateUniqueInput struct {
	Table         string
	Column        string
	Value         interface{}
	IgnoreColumn  string
	IgnoreValue   interface{}
	IncludeColumn string
	IncludeValue  interface{}
	AllowHits     int
}

// ValidateUnique validates that a value is not yet present in a db Table.
func ValidateUnique(db uniqueGetter, input *ValidateUniqueInput) (bool, error) {
	var query string
	var params []interface{}
	var err error

	// Build the unique query depending on ignored values.
	columns := fmt.Sprintf("COUNT(`%s`) as unq", input.Column)
	where := sq.Eq{input.Column: input.Value}

	q := sq.Select(columns).From(input.Table).Where(where).Limit(1)
	if input.IgnoreColumn != "" {
		q = q.Where(sq.NotEq{input.IgnoreColumn: input.IgnoreValue})
	}
	if input.IncludeColumn != "" {
		q = q.Where(sq.Eq{input.IncludeColumn: input.IncludeValue})
	}

	query, params, err = q.ToSql()
	if err != nil {
		return false, errors.WithStack(err)
	}

	// Execute the query.
	type result struct {
		Unq int
	}
	var r result
	err = db.Get(&r, query, params...)
	if err != nil {
		return false, errors.WithStack(err)
	}

	return r.Unq <= input.AllowHits, nil
}
