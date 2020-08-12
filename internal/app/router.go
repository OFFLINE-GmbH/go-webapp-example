package app

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"runtime/debug"
	"time"

	"go-webapp-example/internal/pkg/auth"
	"go-webapp-example/pkg/i18n"
	"go-webapp-example/pkg/log"
	"go-webapp-example/pkg/render"
	"go-webapp-example/pkg/router"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
)

// setupRouter registers all middleware for the application.
func (k *Kernel) setupRouter() {
	k.Router.UseMiddleware(loggerMiddleware(k))
	k.Router.UseMiddleware(middleware.RequestID)
	k.Router.UseMiddleware(recoverer(k.Log))
	k.Router.UseMiddleware(middleware.RealIP)
	k.Router.UseMiddleware(versionHeaderMiddleware)
	k.Router.UseMiddleware(k.Session.Middleware)

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	k.Router.UseMiddleware(corsMiddleware.Handler)
}

// loggerMiddleware logs all requests to the server.
func loggerMiddleware(k *Kernel) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("upgrade") == "websocket" {
				next.ServeHTTP(w, r)
				return
			}
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			t1 := time.Now()
			defer func() {
				duration := time.Since(t1)
				msg := fmt.Sprintf("[%d] %-5s %10s %s from %s", ww.Status(), r.Method, duration, r.RequestURI, r.RemoteAddr)
				if ww.Status() >= 400 {
					k.requestLog.Error(msg)
				} else {
					if duration.Seconds() > 1 {
						k.requestLog.Warn(msg)
					} else {
						k.requestLog.Debug(msg)
					}
				}
			}()
			next.ServeHTTP(ww, r)
		})
	}
}

// setupFrontendRoutes registers all routes that serve static files.
func (k *Kernel) setupFrontendRoutes() {
	k.Router.Group(func(r *router.Mux) {
		r.UseMiddleware(timeoutMiddleware(30 * time.Second))

		// Serve all static frontend files (Vue, React) directly from the static folder.
		r.Handle("/backend/storage/*", http.StripPrefix("/backend/storage/", http.FileServer(http.Dir(k.Config.Server.StorageDir))))
		r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir(k.Config.Server.StaticDir))))

		r.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
			indexPath := filepath.Join(k.Config.Server.StaticDir, "index.html")
			contents, err := ioutil.ReadFile(indexPath)
			if err != nil {
				render.Error(w, err)
				return
			}
			_, err = w.Write(contents)
			if err != nil {
				render.Error(w, err)
			}
		})
	})
}

// setupBackendRoutes registers all backend routes.
func (k *Kernel) setupBackendRoutes() {
	k.Router.Group(func(r *router.Mux) {
		r.UseMiddleware(timeoutMiddleware(30 * time.Second))
		r.UseMiddleware(k.Session.Middleware)

		r.Method(http.MethodPost, "/backend/login", auth.LoginHandler(k.services.User, k.services.Permission, k.services.Audit, k.Locale))
		r.Method(http.MethodPost, "/backend/logout", auth.LogoutHandler(k.Session))
		r.Method(http.MethodGet, "/backend/locale/{locale}", i18n.HandleFunc(k.Config.Server.LocalesDir))
	})
}

// versionHeaderMiddleware adds the current backend version as a HTTP response header.
func versionHeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("X-Webapp-Version", version)
		next.ServeHTTP(w, r)
	})
}

// timeoutMiddleware cancels a request after a certain timeout was reached. It
// ignores websocket requests, they never time out.
func timeoutMiddleware(timeout time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			// Ignore timeouts for websocket requests.
			v := r.Header.Get("Upgrade")
			if v == "websocket" {
				next.ServeHTTP(w, r)
			}

			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer func() {
				cancel()
				if ctx.Err() == context.DeadlineExceeded {
					w.WriteHeader(http.StatusGatewayTimeout)
				}
			}()

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

// recoverer recovers from panics.
func recoverer(logger log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil {
					logger.Errorf("PANIC: %s", rvr)
					logger.Errorf("%+v", rvr)
					logger.Errorf("%+s", debug.Stack())

					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
