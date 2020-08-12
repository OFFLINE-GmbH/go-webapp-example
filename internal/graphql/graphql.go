package graphql

import (
	"context"
	"net/http"
	"runtime/debug"
	"time"

	"go-webapp-example/internal/graphql/gqldataloaders"
	"go-webapp-example/internal/graphql/gqldirectives"
	"go-webapp-example/internal/graphql/gqlresolvers"
	"go-webapp-example/internal/graphql/gqlserver"
	"go-webapp-example/internal/pkg"
	internalauth "go-webapp-example/internal/pkg/auth"
	"go-webapp-example/pkg/auth"
	"go-webapp-example/pkg/i18n"
	"go-webapp-example/pkg/log"
	"go-webapp-example/pkg/session"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/getsentry/sentry-go"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// nolint:gocritic
func New(
	services *pkg.Services,
	sess *session.Store,
	authMngr *auth.Manager,
	logger log.Logger,
	locale *i18n.Locale,
	storageDir string,
) (http.Handler, http.Handler) {
	// authMiddleware is used to authenticate the user and apply directives (like @has)
	authMiddleware := internalauth.Middleware(services.User, sess, logger.WithPrefix("auth.mdlwr"), true)

	// resolver contains all shared dependencies.
	resolver := &gqlresolvers.Resolver{
		Services: services,
		Log:      logger.WithPrefix("graphql"),
		Config:   struct{ StorageDir string }{StorageDir: storageDir},
	}

	c := gqlserver.Config{Resolvers: resolver}
	c.Directives.Restricted = gqldirectives.Restricted(authMngr)

	schema := gqlserver.NewExecutableSchema(c)

	srv := newServer(schema, logger)

	// query is the global GraphQL endpoint each query is sent to.
	query := withMiddleware(
		srv,
		authMiddleware,
		i18n.Middleware(locale),
		gqldataloaders.Middleware(services),
	)
	// playground is used to directly access the graphql api.
	pg := withMiddleware(playground.Handler("GraphQL playground", "/backend/query"), authMiddleware)

	return query, pg
}

// withMiddleware applies multiple middleware to a http.Handler.
func withMiddleware(h http.Handler, middleware ...func(http.Handler) http.Handler) http.Handler {
	for _, mw := range middleware {
		h = mw(h)
	}
	return h
}

func newServer(es graphql.ExecutableSchema, logger log.Logger) *handler.Server {
	srv := handler.New(es)

	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 15 * time.Second,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	})
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})

	srv.SetQueryCache(lru.New(1000))
	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New(100),
	})
	srv.SetRecoverFunc(func(ctx context.Context, err interface{}) (userMessage error) {
		logger.Errorf("%+v", err)
		logger.Errorf("%s", debug.Stack())
		if errString, ok := err.(string); ok {
			sentry.CaptureException(errors.New(errString))
		}
		return errors.New("internal system error")
	})
	srv.SetErrorPresenter(func(ctx context.Context, err error) *gqlerror.Error {
		if gqlErr, isGQLError := err.(*gqlerror.Error); isGQLError {
			// Don't log validation errors.
			if _, isValidation := gqlErr.Extensions["validation"]; !isValidation {
				sentry.CaptureException(err)
				logger.Errorf("%+v", err)
			}
		} else {
			sentry.CaptureException(err)
		}

		return graphql.DefaultErrorPresenter(ctx, err)
	})

	return srv
}
