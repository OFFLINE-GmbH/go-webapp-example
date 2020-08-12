package quote

import (
	"testing"

	"go-webapp-example/internal/graphql/gqlmodels"

	"github.com/stretchr/testify/assert"
)

func TestValidateCreateRequest(t *testing.T) {
	t.Run("Invalid", func(t *testing.T) {
		input := gqlmodels.QuoteInput{
			Author: "",
		}

		err := ValidateCreateRequest(&input)

		assert.True(t, err.Failed())
		assert.Len(t, err.Get("name"), 1)
	})

	t.Run("Valid", func(t *testing.T) {
		input := gqlmodels.QuoteInput{
			Author:  "Test Quote",
			Content: "Test Quote",
		}

		err := ValidateCreateRequest(&input)

		assert.False(t, err.Failed())
		assert.Len(t, err.Errors(), 0)
	})
}
