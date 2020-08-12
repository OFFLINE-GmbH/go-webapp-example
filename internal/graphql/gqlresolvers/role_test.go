package gqlresolvers

import (
	"context"
	"strconv"
	"testing"

	"go-webapp-example/internal/pkg"

	"github.com/99designs/gqlgen/client"
	"github.com/stretchr/testify/assert"
)

type roleFields struct {
	ID   string
	Name string

	Permissions []struct {
		CodeLevel string `json:"code_level"`
	}
}

func TestGraphQL_Role(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	c, services, cleanup := testClient(t)
	defer cleanup()

	t.Run("Roles Query", testRolesQuery(c))
	t.Run("Role Query", testRoleQuery(c))
	t.Run("createRole", testCreateRole(c, services))
	t.Run("updateRole", testUpdateRole(c, services))
	t.Run("deleteRole", testDeleteRole(c, services))
}

func testRolesQuery(c *client.Client) func(t *testing.T) {
	return func(t *testing.T) {
		var resp struct {
			Roles []roleFields
		}

		err := c.Post(`
			query roles {
				  roles {
					id
					name

					permissions { code_level }
			    }
			}`, &resp)

		assert.NoError(t, err)
		assert.Len(t, resp.Roles, 3)

		if len(resp.Roles) >= 2 {
			checkRolesResponse(t, resp.Roles[1])
		}
	}
}

func testRoleQuery(c *client.Client) func(t *testing.T) {
	return func(t *testing.T) {
		var resp struct {
			Role roleFields
		}

		err := c.Post(`
			query Role {
				  role(id: 2) {
					id
					name
					permissions { code_level }
			    }
			}`, &resp)

		assert.NoError(t, err)
		checkRolesResponse(t, resp.Role)
	}
}

func checkRolesResponse(t *testing.T, fields roleFields) {
	assert.Equal(t, fields.ID, "2")
	assert.Equal(t, fields.Name, "reseller")

	assert.Len(t, fields.Permissions, 1)
	if len(fields.Permissions) > 0 {
		assert.Equal(t, fields.Permissions[0].CodeLevel, "test::edit")
	}
}

func testCreateRole(c *client.Client, services *pkg.Services) func(t *testing.T) {
	return func(t *testing.T) {
		var resp struct {
			CreateRole roleFields
		}

		err := c.Post(`
			mutation create {
				  createRole(input: {
					name: "Created"
					permissions: [
						{ code: "test-code", level: "read" } 
					]
				  }) {
					id
					name

					permissions { code_level }
			    }
			}`, &resp)
		assert.NoError(t, err)

		id, _ := strconv.Atoi(resp.CreateRole.ID)
		created, err := services.Role.Find(context.Background(), id)
		assert.NoError(t, err)

		assert.Equal(t, "4", resp.CreateRole.ID)
		assert.Equal(t, "Created", resp.CreateRole.Name)

		assert.Equal(t, created.ID, 4)
		assert.Equal(t, "Created", created.Name)
		assert.NotNil(t, created.CreatedAt)
		assert.NotNil(t, created.UpdatedAt)

		assert.Len(t, resp.CreateRole.Permissions, 1)
		if len(resp.CreateRole.Permissions) > 0 {
			assert.Equal(t, resp.CreateRole.Permissions[0].CodeLevel, "test-code::read")
		}
	}
}

func testUpdateRole(c *client.Client, services *pkg.Services) func(t *testing.T) {
	return func(t *testing.T) {
		var resp struct {
			UpdateRole roleFields
		}

		err := c.Post(`
			mutation update {
				  updateRole(input: {
					id: 1,
					name: "Updated"
					permissions: [
						{ code: "test-code", level: "manage" } 
					]
				  }) {
					id
					name

					permissions { code_level }
			    }
			}`, &resp)
		assert.NoError(t, err)

		updated, err := services.Role.Find(context.Background(), 1)
		assert.NoError(t, err)

		assert.NoError(t, err)
		assert.Equal(t, "1", resp.UpdateRole.ID)
		assert.Equal(t, "Updated", resp.UpdateRole.Name)

		assert.NotEqual(t, updated.ID, 0)
		assert.Equal(t, "Updated", updated.Name)
		assert.NotNil(t, updated.CreatedAt)
		assert.NotNil(t, updated.UpdatedAt)
		assert.NotEqual(t, updated.UpdatedAt, updated.CreatedAt)

		assert.Len(t, resp.UpdateRole.Permissions, 3) // includes read, write and mange
		if len(resp.UpdateRole.Permissions) > 0 {
			var found bool
			for _, permission := range resp.UpdateRole.Permissions {
				if permission.CodeLevel == "test-code::manage" {
					found = true
					break
				}
			}
			assert.Truef(t, found, "permission 'test-code::manage' not found")
		}
	}
}

func testDeleteRole(c *client.Client, services *pkg.Services) func(t *testing.T) {
	return func(t *testing.T) {
		var resp struct {
			DeleteRole []struct {
				ID string
			}
		}

		// Admin role cannot be deleted, error required
		err := c.Post(` mutation delete { deleteRole(id: [1]) { id } }`, &resp)
		assert.Error(t, err)

		c.MustPost(` mutation delete { deleteRole(id: [2]) { id } }`, &resp)

		_, err = services.Role.Find(context.Background(), 2)

		assert.Equal(t, "2", resp.DeleteRole[0].ID)
		assert.Error(t, err)
	}
}
