package user

import (
	"testing"

	"go-webapp-example/internal/graphql/gqlmodels"

	"github.com/stretchr/testify/assert"
)

func TestValidateCreateRequest(t *testing.T) {
	t.Run("Invalid", func(t *testing.T) {
		input := gqlmodels.UserInput{
			Name:           "",
			Password:       "",
			PasswordRepeat: "abc",
		}

		err := ValidateCreateRequest(&input)

		assert.True(t, err.Failed())
		assert.Len(t, err.Get("name"), 1)
		assert.Len(t, err.Get("password"), 1)
		assert.Len(t, err.Get("password_repeat"), 1)
	})

	t.Run("Valid", func(t *testing.T) {
		input := gqlmodels.UserInput{
			Name:           "user",
			Password:       "password",
			PasswordRepeat: "password",
		}

		err := ValidateCreateRequest(&input)

		assert.False(t, err.Failed())
		assert.Len(t, err.Errors(), 0)
	})
}

func TestValidateUpdateRequest(t *testing.T) {
	id := 1

	t.Run("Invalid", func(t *testing.T) {
		input := gqlmodels.UserInput{
			Name:           "",
			Password:       "test",
			PasswordRepeat: "tset",
		}

		err := ValidateUpdateRequest(&input)

		assert.True(t, err.Failed())
		assert.Len(t, err.Get("name"), 1)
		assert.Len(t, err.Get("password"), 0)
		assert.Len(t, err.Get("password_repeat"), 1)
	})

	t.Run("Valid", func(t *testing.T) {
		input := gqlmodels.UserInput{
			ID:             &id,
			Name:           "user",
			Password:       "password",
			PasswordRepeat: "password",
		}

		err := ValidateUpdateRequest(&input)

		assert.False(t, err.Failed())
		assert.Len(t, err.Errors(), 0)
	})

	t.Run("Valid without Password", func(t *testing.T) {
		input := gqlmodels.UserInput{
			ID:             &id,
			Name:           "user",
			Password:       "",
			PasswordRepeat: "",
		}

		err := ValidateUpdateRequest(&input)

		assert.False(t, err.Failed())
		assert.Len(t, err.Errors(), 0)
	})
}
