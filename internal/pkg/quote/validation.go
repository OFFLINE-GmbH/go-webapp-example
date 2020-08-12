package quote

import (
	"go-webapp-example/internal/graphql/gqlmodels"
	"go-webapp-example/pkg/validation"
)

// ValidateCreateRequest validates a create request of this entity.
func ValidateCreateRequest(input *gqlmodels.QuoteInput) *validation.ErrorBag {
	errs := validation.NewErrorBag("quote")

	if input.Author == "" {
		errs.Add("name", "required")
	}
	if input.Content == "" {
		errs.Add("content", "required")
	}

	return errs
}

// ValidateUpdateRequest validates a create request of this entity.
func ValidateUpdateRequest(input *gqlmodels.QuoteInput) *validation.ErrorBag {
	return ValidateCreateRequest(input)
}
