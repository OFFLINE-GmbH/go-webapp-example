package pkg

import (
	"go-webapp-example/internal/pkg/audit"
	"go-webapp-example/internal/pkg/permission"
	"go-webapp-example/internal/pkg/quote"
	"go-webapp-example/internal/pkg/role"
	"go-webapp-example/internal/pkg/user"
	"go-webapp-example/pkg/db"
)

// Services holds all the required services for the graphql server.
type Services struct {
	User       *user.Service
	Role       *role.Service
	Permission *permission.Service
	Audit      *audit.Service
	Quote      *quote.Service
	DB         *db.Connection
}
