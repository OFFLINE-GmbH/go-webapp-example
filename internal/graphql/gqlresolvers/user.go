package gqlresolvers

import (
	"context"

	"go-webapp-example/internal/graphql/gqldataloaders"
	"go-webapp-example/internal/graphql/gqlmodels"
	"go-webapp-example/internal/pkg/entity"
	"go-webapp-example/internal/pkg/user"
	"go-webapp-example/pkg/session"

	"github.com/pkg/errors"
)

type userResolver struct{ *Resolver }

func (r *userResolver) Roles(ctx context.Context, obj *entity.User) ([]*entity.Role, error) {
	return gqldataloaders.CtxLoaders(ctx).RolesByUser.Load(obj.ID)
}
func (r *userResolver) Permissions(ctx context.Context, obj *entity.User) ([]*entity.Permission, error) {
	return gqldataloaders.CtxLoaders(ctx).PermissionsByUser.Load(obj.ID)
}

// Queries

func (r *queryResolver) User(ctx context.Context, id int) (*entity.User, error) {
	return r.Services.User.Find(ctx, id)
}
func (r *queryResolver) Users(ctx context.Context) ([]*entity.User, error) {
	var filtered []*entity.User
	users, err := r.Services.User.Get(ctx)
	if err != nil {
		return filtered, err
	}
	authUser, err := session.UserFromContext(ctx)
	if err != nil {
		return filtered, err
	}
	// Return all users if a super user is logged in. Otherwise, remove
	// other superusers from the result.
	if authUser.IsSuperuser {
		return users, nil
	}
	for _, u := range users {
		if !u.IsSuperuser {
			filtered = append(filtered, u)
		}
	}
	return filtered, nil
}
func (r *queryResolver) AuthUser(ctx context.Context) (*entity.User, error) {
	u, err := session.UserFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get auth user")
	}
	return u, nil
}

// Mutations

func (r *mutationResolver) CreateUser(ctx context.Context, input gqlmodels.UserInput) (*entity.User, error) {
	authUser, err := session.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if !authUser.IsSuperuser && input.IsSuperuser {
		return nil, errors.New("non superuser accounts cannot create superuser accounts")
	}
	if err := user.ValidateCreateRequest(&input); err.Failed() {
		return nil, addErrors(ctx, err)
	}
	u, err := r.Services.User.Create(ctx, toUserEntity(input))
	if err != nil {
		return nil, err
	}
	if input.Roles != nil {
		u, err = r.Services.User.SyncRoles(ctx, u, input.Roles)
		if err != nil {
			return nil, err
		}
	}
	return u, err
}

func (r *mutationResolver) UpdateUser(ctx context.Context, input gqlmodels.UserInput) (*entity.User, error) {
	authUser, err := session.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if !authUser.IsSuperuser && input.IsSuperuser {
		return nil, errors.New("non superuser accounts cannot edit superuser accounts")
	}
	if err := user.ValidateUpdateRequest(&input); err.Failed() {
		return nil, addErrors(ctx, err)
	}
	u, err := r.Services.User.Update(ctx, toUserEntity(input))
	if err != nil {
		return nil, err
	}
	if input.Roles != nil {
		u, err = r.Services.User.SyncRoles(ctx, u, input.Roles)
		if err != nil {
			return nil, err
		}
	}
	return u, err
}

func (r *mutationResolver) DeleteUser(ctx context.Context, ids []int) ([]*entity.User, error) {
	return r.Services.User.Delete(ctx, ids)
}

func toUserEntity(input gqlmodels.UserInput) *entity.User {
	return &entity.User{
		ID:          handleIntPtr(input.ID),
		Name:        input.Name,
		Password:    input.Password,
		IsSuperuser: input.IsSuperuser,
	}
}
