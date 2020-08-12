package app

import (
	"go-webapp-example/internal/graphql"
	"go-webapp-example/pkg/router"
)

// setupGraphQL attaches the /query and /graphql-playground to the application.
func (k *Kernel) setupGraphQL() {
	query, playground := graphql.New(
		k.services,
		k.Session,
		k.Auth,
		k.Log,
		k.Locale,
		k.Config.Server.StorageDir,
	)

	k.Router.Group(func(r *router.Mux) {
		r.Handle("/backend/graphql-playground", playground)
		r.Handle("/backend/query", query)
		// Websocket endpoint. This is only used in dev as in prod
		// the /ws prefix is handled by the proxy container and
		// removed after the appropriate http headers have been set.
		r.Handle("/ws/backend/query", query)
	})
}
