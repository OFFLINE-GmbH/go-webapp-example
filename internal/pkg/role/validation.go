package role

import (
	"go-webapp-example/internal/graphql/gqlmodels"
	"go-webapp-example/pkg/validation"
)

// ValidateCreateRequest validates a create request of this entity.
func ValidateCreateRequest(input *gqlmodels.RoleInput) *validation.ErrorBag {
	errs := validation.NewErrorBag("role")

	if input.Name == "" {
		errs.Add("name", "required")
	}

	return errs
}

// ValidateUpdateRequest validates a create request of this entity.
func ValidateUpdateRequest(input *gqlmodels.RoleInput) *validation.ErrorBag {
	return ValidateCreateRequest(input)
}
