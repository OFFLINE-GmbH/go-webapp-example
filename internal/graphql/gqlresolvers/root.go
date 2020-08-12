//go:generate go run github.com/99designs/gqlgen
package gqlresolvers

import (
	"context"

	"go-webapp-example/internal/graphql/gqlserver"
	"go-webapp-example/internal/pkg"
	"go-webapp-example/pkg/i18n"
	"go-webapp-example/pkg/log"
	"go-webapp-example/pkg/validation"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type Resolver struct {
	Services *pkg.Services
	Log      log.Logger

	Config struct {
		StorageDir string
	}
}

func (r *Resolver) Mutation() gqlserver.MutationResolver {
	return &mutationResolver{r}
}
func (r *Resolver) Query() gqlserver.QueryResolver {
	return &queryResolver{r}
}
func (r *Resolver) User() gqlserver.UserResolver {
	return &userResolver{r}
}
func (r *Resolver) Role() gqlserver.RoleResolver {
	return &roleResolver{r}
}
func (r *Resolver) Permission() gqlserver.PermissionResolver {
	return &permissionResolver{r}
}

type mutationResolver struct{ *Resolver }

type queryResolver struct{ *Resolver }

// addErrors adds errors to the graphql response.
func addErrors(ctx context.Context, data *validation.ErrorBag) error {
	return addErrorsPrefixed(ctx, data, "")
}

// addErrorsPrefixed adds errors to the graphql response with a field prefix.
func addErrorsPrefixed(ctx context.Context, data *validation.ErrorBag, prefix string) error {
	for field, errs := range data.TranslatedErrors(i18n.CtxLocale(ctx)) {
		for _, err := range errs {
			graphql.AddError(ctx, &gqlerror.Error{
				Message: err.Message,
				Extensions: map[string]interface{}{
					"field":      prefix + field,
					"data":       err.Data,
					"validation": true,
				},
			})
		}
	}
	return validation.ErrFailed
}

// handleIntPtr returns 0 for a nil pointer value, otherwise returns the original input.
func handleIntPtr(id *int) int {
	if id == nil {
		return 0
	}
	return *id
}
