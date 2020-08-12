package gqlresolvers

import (
	"context"

	"go-webapp-example/internal/graphql/gqldataloaders"
	"go-webapp-example/internal/graphql/gqlmodels"
	"go-webapp-example/internal/pkg/entity"
	"go-webapp-example/internal/pkg/role"
	"go-webapp-example/pkg/session"
)

type roleResolver struct{ *Resolver }

func (r *roleResolver) Users(ctx context.Context, obj *entity.Role) ([]*entity.User, error) {
	return gqldataloaders.CtxLoaders(ctx).UsersByRole.Load(obj.ID)
}

func (r *roleResolver) Permissions(ctx context.Context, obj *entity.Role) ([]*entity.Permission, error) {
	return gqldataloaders.CtxLoaders(ctx).PermissionsByRole.Load(obj.ID)
}

type permissionResolver struct{ *Resolver }

func (r *permissionResolver) Level(ctx context.Context, obj *entity.Permission) (string, error) {
	return string(obj.Level), nil
}

// Queries

func (r *queryResolver) Roles(ctx context.Context) ([]*entity.Role, error) {
	var filtered []*entity.Role
	roles, err := r.Services.Role.Get(ctx)
	if err != nil {
		return filtered, err
	}
	authUser, err := session.UserFromContext(ctx)
	if err != nil {
		return filtered, err
	}
	// Return all roles if a super user is logged in. Otherwise, remove
	// other admin rule from the result.
	if authUser.IsSuperuser {
		return roles, nil
	}
	for _, u := range roles {
		if u.ID > 1 {
			filtered = append(filtered, u)
		}
	}
	return filtered, nil
}

func (r *queryResolver) Role(ctx context.Context, id int) (*entity.Role, error) {
	return r.Services.Role.Find(ctx, id)
}

// Mutations

func (r *mutationResolver) CreateRole(ctx context.Context, input gqlmodels.RoleInput) (*entity.Role, error) {
	if err := role.ValidateCreateRequest(&input); err.Failed() {
		return nil, addErrors(ctx, err)
	}
	u, err := r.Services.Role.Create(ctx, toRoleEntity(input))
	if err != nil {
		return nil, err
	}
	if input.Permissions != nil {
		u, err = r.Services.Role.SyncPermissions(ctx, u, mapPermissions(input.Permissions))
		if err != nil {
			return nil, err
		}
	}
	if input.Users != nil {
		u, err = r.Services.Role.SyncUsers(ctx, u, input.Users)
	}
	return u, err
}

func (r *mutationResolver) UpdateRole(ctx context.Context, input gqlmodels.RoleInput) (*entity.Role, error) {
	if err := role.ValidateUpdateRequest(&input); err.Failed() {
		return nil, addErrors(ctx, err)
	}
	u, err := r.Services.Role.Update(ctx, toRoleEntity(input))
	if err != nil {
		return nil, err
	}
	if input.Permissions != nil {
		u, err = r.Services.Role.SyncPermissions(ctx, u, mapPermissions(input.Permissions))
		if err != nil {
			return nil, err
		}
	}
	if input.Permissions != nil {
		u, err = r.Services.Role.SyncUsers(ctx, u, input.Users)
		if err != nil {
			return nil, err
		}
	}
	return u, err
}

func (r *mutationResolver) DeleteRole(ctx context.Context, ids []int) ([]*entity.Role, error) {
	return r.Services.Role.Delete(ctx, ids)
}

func toRoleEntity(input gqlmodels.RoleInput) *entity.Role {
	return &entity.Role{
		ID:   handleIntPtr(input.ID),
		Name: input.Name,
	}
}

// mapPermissions turns a DisplayGroupInput into a map of SectionIDs to CallTypeIDs
func mapPermissions(input []*gqlmodels.PermissionInput) []*entity.Permission {
	var permissions []*entity.Permission
	for _, p := range input {
		permissions = append(permissions, &entity.Permission{
			Code:  p.Code,
			Level: entity.PermissionLevel(p.Level),
		})
	}
	return permissions
}
