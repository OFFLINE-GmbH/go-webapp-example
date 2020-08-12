package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// nolint:scopelint
func TestValidateIP(t *testing.T) {
	tt := []struct {
		input string
		want  bool
	}{
		{input: "192.168.1.1", want: true},
		{input: "10.0.0.1", want: true},
		{input: "255.255.255.255", want: true},
		{input: "300.300.300.300", want: false},
		{input: "127.0.0.1", want: true},
		{input: "0.0.0.0", want: true},
		{input: "127.0", want: false},
		{input: "127.0.0.0.1", want: false},
	}

	for _, tc := range tt {
		t.Run(fmt.Sprintf("ValidateIP: %s", tc.input), func(t *testing.T) {
			got := ValidateIP(tc.input)
			if got != tc.want {
				t.Errorf("ValidateIP(%s), got %t; want %t", tc.input, got, tc.want)
			}
		})
	}
}

type mockDB struct {
	query string
	args  []interface{}
}

func (db *mockDB) Get(target interface{}, query string, args ...interface{}) error {
	db.query = query
	db.args = args

	return nil
}

// nolint:scopelint,funlen
func TestValidateUnique(t *testing.T) {
	db := mockDB{}

	tt := []struct {
		name      string
		input     *ValidateUniqueInput
		wantQuery string
		wantArgs  []interface{}
	}{
		{
			name: "Simple Unique",
			input: &ValidateUniqueInput{
				Table:  "atable",
				Column: "acolumn",
				Value:  5,
			},
			wantQuery: "SELECT COUNT(`acolumn`) as unq FROM atable WHERE acolumn = ? LIMIT 1",
			wantArgs:  []interface{}{5},
		},
		{
			name: "Ignore Unique",
			input: &ValidateUniqueInput{
				Table:        "atable",
				Column:       "acolumn",
				Value:        5,
				IgnoreValue:  3,
				IgnoreColumn: "id",
			},
			wantQuery: "SELECT COUNT(`acolumn`) as unq FROM atable WHERE acolumn = ? AND id <> ? LIMIT 1",
			wantArgs:  []interface{}{5, 3},
		},
		{
			name: "Include Unique",
			input: &ValidateUniqueInput{
				Table:         "atable",
				Column:        "acolumn",
				Value:         5,
				IncludeValue:  4,
				IncludeColumn: "anothercolumn",
			},
			wantQuery: "SELECT COUNT(`acolumn`) as unq FROM atable WHERE acolumn = ? AND anothercolumn = ? LIMIT 1",
			wantArgs:  []interface{}{5, 4},
		},
		{
			name: "Ignore Include Unique",
			input: &ValidateUniqueInput{
				Table:         "atable",
				Column:        "acolumn",
				Value:         5,
				IgnoreValue:   3,
				IgnoreColumn:  "id",
				IncludeValue:  4,
				IncludeColumn: "anothercolumn",
			},
			wantQuery: "SELECT COUNT(`acolumn`) as unq FROM atable WHERE acolumn = ? AND id <> ? AND anothercolumn = ? LIMIT 1",
			wantArgs:  []interface{}{5, 3, 4},
		},
	}

	for _, tc := range tt {
		t.Run(fmt.Sprintf("ValidateUnique: %s", tc.name), func(t *testing.T) {
			_, err := ValidateUnique(&db, tc.input)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantQuery, db.query)
			assert.Equal(t, tc.wantArgs, db.args)
		})
	}
}
