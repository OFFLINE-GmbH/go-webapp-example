//go:generate go run github.com/vektah/dataloaden RoleSliceLoader int []*go-webapp-example/internal/pkg/entity.Role
//go:generate go run github.com/vektah/dataloaden UserSliceLoader int []*go-webapp-example/internal/pkg/entity.User
//go:generate go run github.com/vektah/dataloaden PermissionSliceLoader int []*go-webapp-example/internal/pkg/entity.Permission
package gqldataloaders

import (
	"context"
	"net/http"
	"time"

	"go-webapp-example/internal/pkg"
	"go-webapp-example/internal/pkg/entity"
)

type ctxKeyType struct{ name string }

var ctxKey = ctxKeyType{"userCtx"}

type Loaders struct {
	RolesByUser       *RoleSliceLoader
	PermissionsByUser *PermissionSliceLoader
	PermissionsByRole *PermissionSliceLoader
	UsersByRole       *UserSliceLoader
}

func Middleware(services *pkg.Services) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			dlCtx := ContextWithLoaders(r.Context(), services)
			next.ServeHTTP(w, r.WithContext(dlCtx))
		})
	}
}

// nolint:funlen
func ContextWithLoaders(
	ctx context.Context,
	services *pkg.Services,
) context.Context {
	ldrs := Loaders{}
	wait := 200 * time.Microsecond

	// Fetch all roles for a given slice of user ids.
	ldrs.RolesByUser = &RoleSliceLoader{
		maxBatch: 100,
		wait:     wait,
		fetch: func(ids []int) ([][]*entity.Role, []error) {
			result := make([][]*entity.Role, len(ids))
			items, err := services.Role.GetByUserID(ctx, ids)
			if err != nil {
				return result, []error{err}
			}
			for i, key := range ids {
				result[i] = items[key]
			}
			return result, nil
		},
	}

	// Fetch all permissions for a given slice of user ids.
	ldrs.PermissionsByUser = &PermissionSliceLoader{
		maxBatch: 100,
		wait:     wait,
		fetch: func(ids []int) ([][]*entity.Permission, []error) {
			result := make([][]*entity.Permission, len(ids))
			items, err := services.Permission.GetByUserID(ctx, ids)
			if err != nil {
				return result, []error{err}
			}
			for i, key := range ids {
				result[i] = items[key]
			}
			return result, nil
		},
	}

	// Fetch all permissions for a given slice of role ids.
	ldrs.PermissionsByRole = &PermissionSliceLoader{
		maxBatch: 100,
		wait:     wait,
		fetch: func(ids []int) ([][]*entity.Permission, []error) {
			result := make([][]*entity.Permission, len(ids))
			items, err := services.Permission.GetByRoleID(ctx, ids)
			if err != nil {
				return result, []error{err}
			}
			for i, key := range ids {
				result[i] = items[key]
			}
			return result, nil
		},
	}

	// Fetch all users for a given slice of role ids.
	ldrs.UsersByRole = &UserSliceLoader{
		maxBatch: 100,
		wait:     wait,
		fetch: func(ids []int) ([][]*entity.User, []error) {
			result := make([][]*entity.User, len(ids))
			items, err := services.User.GetByRoleID(ctx, ids)
			if err != nil {
				return result, []error{err}
			}
			for i, key := range ids {
				result[i] = items[key]
			}
			return result, nil
		},
	}

	return context.WithValue(ctx, ctxKey, ldrs)
}

func CtxLoaders(ctx context.Context) Loaders {
	return ctx.Value(ctxKey).(Loaders)
}
