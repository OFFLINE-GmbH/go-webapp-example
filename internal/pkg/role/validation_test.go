package role

import (
	"testing"

	"go-webapp-example/internal/graphql/gqlmodels"

	"github.com/stretchr/testify/assert"
)

func TestValidateCreateRequest(t *testing.T) {
	t.Run("Invalid", func(t *testing.T) {
		input := gqlmodels.RoleInput{
			Name: "",
		}

		err := ValidateCreateRequest(&input)

		assert.True(t, err.Failed())
		assert.Len(t, err.Get("name"), 1)
	})

	t.Run("Valid", func(t *testing.T) {
		input := gqlmodels.RoleInput{
			Name: "user",
		}

		err := ValidateCreateRequest(&input)

		assert.False(t, err.Failed())
		assert.Len(t, err.Errors(), 0)
	})
}

func TestValidateUpdateRequest(t *testing.T) {
	id := 1

	t.Run("Invalid", func(t *testing.T) {
		input := gqlmodels.RoleInput{
			Name: "",
		}

		err := ValidateUpdateRequest(&input)

		assert.True(t, err.Failed())
		assert.Len(t, err.Get("name"), 1)
	})

	t.Run("Valid", func(t *testing.T) {
		input := gqlmodels.RoleInput{
			ID:   &id,
			Name: "user",
		}

		err := ValidateUpdateRequest(&input)

		assert.False(t, err.Failed())
		assert.Len(t, err.Errors(), 0)
	})
}
