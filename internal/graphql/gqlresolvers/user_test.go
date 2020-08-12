package gqlresolvers

import (
	"context"
	"strconv"
	"testing"

	"go-webapp-example/internal/pkg"

	"github.com/99designs/gqlgen/client"
	"github.com/stretchr/testify/assert"
)

func TestGraphQL_User(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	c, services, cleanup := testClient(t)
	defer cleanup()

	t.Run("Users Query", testUsersQuery(c))
	t.Run("authUser Query", testAuthUserQuery(c))
	t.Run("createUser", testCreateUser(c, services))
	t.Run("updateUser", testUpdateUser(c, services))
	t.Run("deleteUser", testDeleteUser(c, services))
}

func testUsersQuery(c *client.Client) func(t *testing.T) {
	return func(t *testing.T) {
		var resp struct {
			Users []struct {
				ID          string `json:"id"`
				Name        string
				IsSuperuser bool `json:"is_superuser"`
				Roles       []struct {
					ID   string
					Name string
				}
				Permissions []struct {
					Code  string
					Level string
				}
			}
		}

		c.MustPost(`
			query Users {
				users {
					id
					name
					is_superuser
					roles {
						id
						name
					}
					permissions {
						code
						level
					}
				}
			}`, &resp)

		assert.Equal(t, resp.Users[0].ID, "1")
		assert.Equal(t, resp.Users[0].Name, "admin")
		assert.Equal(t, resp.Users[0].IsSuperuser, true)
		assert.Equal(t, resp.Users[0].Roles[0].ID, "1")
		assert.Equal(t, resp.Users[0].Permissions[0].Code, "device")
		assert.Equal(t, resp.Users[0].Permissions[0].Level, "edit")
	}
}

func testAuthUserQuery(c *client.Client) func(t *testing.T) {
	return func(t *testing.T) {
		var resp struct {
			AuthUser struct {
				ID          string `json:"id"`
				Name        string
				IsSuperuser bool `json:"is_superuser"`
				Roles       []struct {
					ID   string
					Name string
				}
				Permissions []struct {
					Code string
				}
			}
		}

		c.MustPost(`
			query authUser {
				authUser {
					id
					name
					is_superuser
					roles {
						id
						name
					}
					permissions {
						code
					}
				}
			}`, &resp)

		assert.Equal(t, resp.AuthUser.ID, "1")
		assert.Equal(t, resp.AuthUser.Name, "admin")
		assert.Equal(t, resp.AuthUser.IsSuperuser, true)
		assert.Equal(t, resp.AuthUser.Roles[0].ID, "1")
		assert.Equal(t, resp.AuthUser.Permissions[0].Code, "device")
	}
}

func testCreateUser(c *client.Client, services *pkg.Services) func(t *testing.T) {
	return func(t *testing.T) {
		var resp struct {
			CreateUser struct {
				ID          string `json:"id"`
				Name        string
				IsSuperuser bool `json:"is_superuser"`
			}
		}

		c.MustPost(`
			mutation {
				createUser(input: {name: "admin", password: "1234", password_repeat: "1234", is_superuser: true, roles: []}) {
					id
 					name
                    is_superuser
				}
			}`, &resp)

		id, _ := strconv.Atoi(resp.CreateUser.ID)
		created, err := services.User.Find(context.Background(), id)
		assert.NoError(t, err)

		assert.Equal(t, resp.CreateUser.Name, "admin")
		assert.NotEqual(t, created.ID, 0)
		assert.Equal(t, created.Name, "admin")
		assert.Equal(t, created.IsSuperuser, true)
	}
}

func testUpdateUser(c *client.Client, services *pkg.Services) func(t *testing.T) {
	return func(t *testing.T) {
		var resp struct {
			UpdateUser struct {
				ID          string
				Name        string
				IsSuperuser bool `json:"is_superuser"`
			}
		}

		c.MustPost(`
			mutation {
				  updateUser(input: {
					id: 1,
					name: "New Name",
					password: "",
					password_repeat: "",
					is_superuser: false,
					roles: [],
				  }) {
					id
					name
					is_superuser
			    }
			}`, &resp)

		updated, err := services.User.Find(context.Background(), 1)
		assert.NoError(t, err)

		assert.Equal(t, "1", resp.UpdateUser.ID)
		assert.Equal(t, "New Name", resp.UpdateUser.Name)
		assert.Equal(t, 1, updated.ID)
		assert.Equal(t, "New Name", updated.Name)
		assert.Equal(t, false, updated.IsSuperuser)
	}
}

func testDeleteUser(c *client.Client, services *pkg.Services) func(t *testing.T) {
	return func(t *testing.T) {
		var resp struct {
			DeleteUser []struct {
				ID string
			}
		}

		// Admin cannot be deleted, error required.
		err := c.Post(` mutation { deleteUser(id: [1]) { id } }`, &resp)
		assert.Error(t, err)

		c.MustPost(` mutation { deleteUser(id: [2]) { id } }`, &resp)

		_, err = services.User.Find(context.Background(), 2)

		assert.Equal(t, "2", resp.DeleteUser[0].ID)
		assert.Error(t, err)
	}
}
