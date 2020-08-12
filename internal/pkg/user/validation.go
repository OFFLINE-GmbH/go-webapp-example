package user

import (
	"go-webapp-example/internal/graphql/gqlmodels"
	"go-webapp-example/pkg/validation"
)

// ValidateCreateRequest validates a create request of this entity.
func ValidateCreateRequest(input *gqlmodels.UserInput) *validation.ErrorBag {
	errs := validation.NewErrorBag("user")

	if input.Name == "" {
		errs.Add("name", "required")
	}
	if input.Password == "" {
		errs.Add("password", "required")
	} else if len(input.Password) < 4 {
		errs.AddData("password", "min", map[string]string{"size": "4"})
	}
	if input.PasswordRepeat == "" {
		errs.Add("password_repeat", "required")
	}
	if input.Password != input.PasswordRepeat {
		errs.Add("password_repeat", "no_match")
	}

	return errs
}

// ValidateUpdateRequest validates a create request of this entity.
func ValidateUpdateRequest(input *gqlmodels.UserInput) *validation.ErrorBag {
	errs := validation.NewErrorBag("user")

	if input.ID == nil || *input.ID < 1 {
		errs.Add("id", "required")
	}

	if input.Name == "" {
		errs.Add("name", "required")
	}

	if input.Password != "" {
		if len(input.Password) < 4 {
			errs.AddData("password", "min_length", map[string]string{"size": "4"})
		}
		if input.PasswordRepeat == "" {
			errs.Add("password_repeat", "required")
		}
		if input.Password != input.PasswordRepeat {
			errs.Add("password_repeat", "no_match")
		}
	}

	return errs
}

// ValidateAuthRequest validates a authentication request.
func ValidateAuthRequest(username, password string) *validation.ErrorBag {
	errs := validation.NewErrorBag("user")

	if username == "" {
		errs.Add("username", "required")
	}

	if password == "" {
		errs.Add("password", "required")
	} else if len(password) < 4 {
		errs.AddData("password", "min_length", map[string]string{"size": "4"})
	}

	return errs
}
