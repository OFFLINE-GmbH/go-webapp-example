package gqlresolvers

import (
	"context"
	"strconv"
	"testing"

	"go-webapp-example/internal/pkg"

	"github.com/99designs/gqlgen/client"
	"github.com/stretchr/testify/assert"
)

type quoteFields struct {
	ID      string
	Author  string `json:"author"`
	Content string `json:"content"`
}

func TestGraphQL_Quote(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	c, services, cleanup := testClient(t)
	defer cleanup()

	t.Run("Quotes Query", testQuotesQuery(c))
	t.Run("Quote Query", testQuoteQuery(c))
	t.Run("createQuote", testCreateQuote(c, services))
	t.Run("updateQuote", testUpdateQuote(c, services))
	t.Run("deleteQuote", testDeleteQuote(c, services))
}

func testQuotesQuery(c *client.Client) func(t *testing.T) {
	return func(t *testing.T) {
		var resp struct {
			Quotes []quoteFields
		}

		err := c.Post(`
			query quotes {
				  quotes {
					id
					author
					content
			    }
			}`, &resp)

		assert.NoError(t, err)
		assert.Len(t, resp.Quotes, 2)

		if len(resp.Quotes) > 0 {
			checkQuotesResponse(t, resp.Quotes[0])
		}
	}
}

func testQuoteQuery(c *client.Client) func(t *testing.T) {
	return func(t *testing.T) {
		var resp struct {
			Quote quoteFields
		}

		err := c.Post(`
			query Quote {
				  quote(id: 1) {
					id
					author
					content
			    }
			}`, &resp)

		assert.NoError(t, err)
		checkQuotesResponse(t, resp.Quote)
	}
}

func checkQuotesResponse(t *testing.T, fields quoteFields) {
	assert.Equal(t, "1", fields.ID)
	assert.Equal(t, "Quote text", fields.Content)
	assert.Equal(t, "A author", fields.Author)
}

func testCreateQuote(c *client.Client, services *pkg.Services) func(t *testing.T) {
	return func(t *testing.T) {
		var resp struct {
			CreateQuote quoteFields
		}

		err := c.Post(`
			mutation create {
				  createQuote(input: {
					author: "Test"
					content: "Content"
				  }) {
					id
			    }
			}`, &resp)
		assert.NoError(t, err)

		id, _ := strconv.Atoi(resp.CreateQuote.ID)
		created, err := services.Quote.Find(context.Background(), id)
		assert.NoError(t, err)

		assert.NotEqual(t, "0", resp.CreateQuote.ID)

		assert.NotEqual(t, created.ID, 0)
		assert.Equal(t, "Test", created.Author)
		assert.Equal(t, "Content", created.Content)
		assert.NotNil(t, created.CreatedAt)
		assert.NotNil(t, created.UpdatedAt)
	}
}

func testUpdateQuote(c *client.Client, services *pkg.Services) func(t *testing.T) {
	return func(t *testing.T) {
		var resp struct {
			UpdateQuote quoteFields
		}

		err := c.Post(`
			mutation update {
				  updateQuote(input: {
					id: 1,
					content: "Updated"
					author: "Author"
				  }) {
					id
			    }
			}`, &resp)
		assert.NoError(t, err)

		updated, err := services.Quote.Find(context.Background(), 1)
		assert.NoError(t, err)

		assert.NoError(t, err)
		assert.Equal(t, "1", resp.UpdateQuote.ID)

		assert.NotEqual(t, updated.ID, 0)
		assert.Equal(t, "Updated", updated.Content)
		assert.Equal(t, "Author", updated.Author)
		assert.NotNil(t, updated.CreatedAt)
		assert.NotNil(t, updated.UpdatedAt)
		assert.NotEqual(t, updated.UpdatedAt, updated.CreatedAt)
	}
}

func testDeleteQuote(c *client.Client, services *pkg.Services) func(t *testing.T) {
	return func(t *testing.T) {
		var resp struct {
			DeleteQuote []struct {
				ID string
			}
		}

		c.MustPost(`
			mutation delete {
				  deleteQuote(id: [1]) {
					id
			    }
			}`, &resp)

		_, err := services.Quote.Find(context.Background(), 1)

		assert.Equal(t, "1", resp.DeleteQuote[0].ID)
		assert.Error(t, err)
	}
}
