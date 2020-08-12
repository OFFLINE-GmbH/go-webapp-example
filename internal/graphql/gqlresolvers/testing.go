package gqlresolvers

import (
	"context"
	"net/http"
	"testing"

	"go-webapp-example/internal/graphql/gqldataloaders"
	"go-webapp-example/internal/graphql/gqlserver"
	"go-webapp-example/internal/pkg"
	"go-webapp-example/internal/pkg/audit"
	"go-webapp-example/internal/pkg/entity"
	"go-webapp-example/internal/pkg/permission"
	"go-webapp-example/internal/pkg/quote"
	"go-webapp-example/internal/pkg/role"
	"go-webapp-example/internal/pkg/test"
	"go-webapp-example/internal/pkg/user"
	"go-webapp-example/pkg/auth"
	"go-webapp-example/pkg/i18n"
	"go-webapp-example/pkg/log"
	"go-webapp-example/pkg/session"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
)

// testClient returns a graphql client for testing.
// nolint:funlen
func testClient(t *testing.T) (cli *client.Client, services *pkg.Services, cleanup func()) {
	db, cleanup := test.DB(t)

	logger := log.NewNullLogger()
	authManager, err := auth.New(db, logger)
	if err != nil {
		t.Fatalf("failed to load auth manager: %s", err)
	}
	authManager.AddRoleForUser(1, 1)
	authManager.AddRolePermission(1, "quote", "edit")
	authManager.AddRolePermission(2, "test", "edit")

	sess := session.New(db.Connection())

	auditor := audit.NewService(audit.NewStore(db), logger)

	services = &pkg.Services{
		DB:         db,
		User:       user.NewService(user.NewStore(db, authManager, auditor), sess),
		Role:       role.NewService(role.NewStore(db, authManager)),
		Permission: permission.NewService(permission.NewStore(db, authManager)),
		Quote:      quote.NewService(quote.NewStore(db, auditor)),
		Audit:      auditor,
	}

	// resolver contains all shared dependencies.
	resolver := &Resolver{
		Services: services,
		Log:      logger,
	}
	// Build the graphql config.
	c := gqlserver.Config{Resolvers: resolver}
	// Add directives.
	c.Directives.Restricted = func(ctx context.Context, obj interface{}, next graphql.Resolver, permissions []string) (res interface{}, err error) {
		return next(ctx)
	}

	schema := gqlserver.NewExecutableSchema(c)

	query := withMiddleware(
		handler.NewDefaultServer(schema),
		authMiddleware,
		i18n.Middleware(&i18n.Locale{}),
		gqldataloaders.Middleware(services),
	)

	return client.New(query), services, cleanup
}

// authMiddleware injects a test admin user into the request context.
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), session.CtxKey, &entity.User{Name: "admin", ID: 1, IsSuperuser: true})

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

// withMiddleware applies multiple middleware to a http.Handler.
func withMiddleware(h http.Handler, middleware ...func(http.Handler) http.Handler) http.Handler {
	for _, mw := range middleware {
		h = mw(h)
	}
	return h
}
