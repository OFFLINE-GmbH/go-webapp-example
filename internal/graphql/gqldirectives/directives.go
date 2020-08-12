package gqldirectives

import (
	"context"
	"strings"

	"go-webapp-example/pkg/auth"
	"go-webapp-example/pkg/session"

	"github.com/99designs/gqlgen/graphql"
	"github.com/pkg/errors"
)

const (
	// Error codes that are using in the frontend.
	ErrMissingAuth       = "MISSING_AUTH"
	ErrMissingPermission = "MISSING_PERMISSION"
)

// RestrictedFn is the "has" directive function.
type RestrictedFn func(ctx context.Context, obj interface{}, next graphql.Resolver, permissions []string) (res interface{}, err error)

// Restricted checks if the currently authenticated user has a certain permission.
func Restricted(a *auth.Manager) RestrictedFn {
	return func(ctx context.Context, obj interface{}, next graphql.Resolver, permissions []string) (res interface{}, err error) {
		u, err := session.UserFromContext(ctx)
		if err != nil {
			return nil, errors.Errorf("auth check: %s: %v", ErrMissingAuth, err)
		}

		// superusers can do everything.
		if u.IsSuperuser {
			return next(ctx)
		}

		for _, p := range permissions {
			parts := strings.Split(p, "::")
			if len(parts) != 2 {
				return nil, errors.Errorf("auth check: invalid permission code, format \"permission.code::level\" expected: %v", p)
			}
			if !a.Can(u.ID, parts[0], parts[1]) {
				return nil, errors.Errorf("auth check: %s: %v", ErrMissingPermission, p)
			}
		}

		return next(ctx)
	}
}
