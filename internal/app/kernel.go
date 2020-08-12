package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go-webapp-example/internal/pkg"
	"go-webapp-example/internal/pkg/audit"
	"go-webapp-example/internal/pkg/permission"
	"go-webapp-example/internal/pkg/quote"
	"go-webapp-example/internal/pkg/role"
	"go-webapp-example/internal/pkg/user"
	"go-webapp-example/pkg/auth"
	"go-webapp-example/pkg/cache"
	"go-webapp-example/pkg/db"
	"go-webapp-example/pkg/i18n"
	"go-webapp-example/pkg/log"
	"go-webapp-example/pkg/router"
	"go-webapp-example/pkg/session"

	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
)

// Possible application states.
const (
	StateStopped int = iota
	StateStarting
	StateRunning
	StateStopping
)

// version is set during build time.
var version = "develop"

// Kernel contains all shared application dependencies.
type Kernel struct {
	// server holds a pointer to the http server.
	server *http.Server
	// Log is the global application logger.
	Log log.Logger
	// Config contains the applications's configuration.
	Config *Config
	// DB is the globally shared database connection.
	DB *db.Connection
	// Session is the globally shared session store.
	Session *session.Store
	// Cache is the application's cache store.
	Cache *cache.Store
	// Auth is the applications' auth manager.
	Auth *auth.Manager
	// Router is the applications HTTP mux.
	Router *router.Mux
	// Locale is used to translate messages
	Locale *i18n.Locale

	// state is the current state the application is in.
	state int
	// services contains all the applications' services.
	services *pkg.Services
	// wg allows processes to register themselves for a graceful shutdown.
	wg *sync.WaitGroup
	// context holds the application wide context.
	context cancelContext
	// requestLog is used to log incoming http requests.
	requestLog log.Logger
	// shutdownFns are run at shutdown.
	shutdownFns []func() error
}

// cancelContext is a context with a cancel function.
type cancelContext struct {
	cancel context.CancelFunc
	ctx    context.Context
}

// New returns a new app kernel instance.
// nolint:funlen
func New() (*Kernel, error) {
	config := LoadConfig()
	logger := log.New(os.Stderr, config.Log.Level, config.Log.Dir)

	database, err := db.New("mysql", config.Database.DSN(), logger.WithPrefix("db"))
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to db")
	}

	authManager, err := auth.New(database, logger.WithPrefix("auth"))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create auth manager")
	}

	sess := session.New(database.Connection())

	locale, err := i18n.FromFiles(config.Server.LocalesDir, config.App.Locale)
	if err != nil {
		// In case of error, proceed anyway.
		logger.Errorf("failed to load locales: %s", err)
	}

	// Create a global application context.
	ctx, cancel := context.WithCancel(context.Background())

	// Build the Kernel struct with all dependencies.
	app := &Kernel{
		state:   StateStarting,
		Log:     logger,
		Config:  config,
		DB:      database,
		Session: sess,
		Cache:   cache.New(),
		Auth:    authManager,
		Router:  router.New(),
		Locale:  locale,

		wg:      &sync.WaitGroup{},
		context: cancelContext{cancel: cancel, ctx: ctx},

		services: &pkg.Services{},
	}

	err = app.setupLogger()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	app.setupServices()
	app.setupAuth()
	app.setupRouter()
	app.setupBackendRoutes()
	app.setupFrontendRoutes()
	app.setupGraphQL()

	app.state = StateRunning

	return app, nil
}

// setupLogger registers all loggers.
func (k *Kernel) setupLogger() error {
	// Request log
	requestW, requestCleanup, err := log.GetMaxLatencyWriter(filepath.Join(k.Config.Log.Dir, "access.log"))
	if err != nil {
		return errors.WithStack(err)
	}
	k.registerShutdownFn(func() error {
		requestCleanup()
		return nil
	})
	k.requestLog = log.NewFromWriter(requestW)
	return nil
}

// setupServices registers all services.
func (k *Kernel) setupServices() {
	k.services.DB = k.DB

	k.services.Audit = audit.NewService(audit.NewStore(k.DB), k.Log.WithPrefix("audit"))
	k.services.Quote = quote.NewService(quote.NewStore(k.DB, k.services.Audit))
	k.services.User = user.NewService(user.NewStore(k.DB, k.Auth, k.services.Audit), k.Session)
	k.services.Role = role.NewService(role.NewStore(k.DB, k.Auth))
	k.services.Permission = permission.NewService(permission.NewStore(k.DB, k.Auth))
}

// ServeHTTP serves the app using the registered router.
func (k *Kernel) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	k.Router.ServeHTTP(w, r)
}

// Listen starts the http server.
func (k *Kernel) Listen(server *http.Server) {
	var err error
	k.server = server

	logger := k.Log.WithPrefix("http.server")

	logger.WithFields(log.Fields{"port": server.Addr}).Infof("staring http server")

	listen := func() error {
		if err = server.ListenAndServe(); err != http.ErrServerClosed {
			msg := fmt.Sprintf("failed to start http server: %v", err)
			sentry.CaptureException(errors.New(msg))
			logger.Error(msg)
			return err
		}
		return nil
	}

	for {
		// If the app is stopped or stopping, don't retry to start the server.
		if k.state == StateStopping || k.state == StateStopped {
			logger.Tracef("skipping restarts of server because app is not in running state: state is %d", k.state)
			return
		}

		if err = listen(); err != nil {
			time.Sleep(2 * time.Second)
			logger.Infof("re-staring http server after error on: %v", server.Addr)
			continue
		}
		return
	}
}

// Shutdown stops the application.
func (k *Kernel) Shutdown(ctx context.Context) error {
	if k.state != StateRunning {
		k.Log.WithPrefix("app").Warn("Application cannot be shutdown since current state is not 'running'")
		return nil
	}

	// Make sure Sentry sends all buffered data before the shutdown.
	defer sentry.Flush(2 * time.Second)

	k.state = StateStopping
	defer func() {
		k.state = StateStopped
	}()

	if k.server != nil {
		if err := k.server.Shutdown(ctx); err != nil {
			k.Log.Errorf("server shutdown error: %v\n", err)
		} else {
			k.Log.Infoln("HTTP server stopped")
		}
	}

	// Cancel global context, then wait for all processes to quit.
	k.context.cancel()
	done := make(chan struct{})
	go func() {
		k.wg.Wait()
		close(done)
	}()

	// Run shutdown functions.
	for _, fn := range k.shutdownFns {
		shutdownErr := fn()
		if shutdownErr != nil {
			k.Log.Errorf("shutdown function returned error: %v\n", shutdownErr)
		}
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
	}

	return k.DB.Close()
}

// Migrate runs all DB and upgrade migrations.
func (k *Kernel) Migrate() error {
	m, err := db.NewMigrator(k.DB, k.Log, k.Config.Database.Migrations)
	if err != nil {
		return errors.WithStack(err)
	}
	if err = m.Up(); err != nil {
		return errors.WithStack(err)
	}

	upd := NewUpdater(k)
	err = upd.Run(k.context.ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// registerShutdownFn registers a function to be run at shutdown.
func (k *Kernel) registerShutdownFn(fn func() error) {
	k.shutdownFns = append(k.shutdownFns, fn)
}

// Version returns the app's version.
func Version() string {
	return version
}
