package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go-webapp-example/internal/pkg/audit"
	"go-webapp-example/internal/pkg/entity"
	"go-webapp-example/internal/pkg/permission"
	"go-webapp-example/internal/pkg/user"
	"go-webapp-example/pkg/i18n"
	"go-webapp-example/pkg/log"
	"go-webapp-example/pkg/render"
	"go-webapp-example/pkg/session"
	"go-webapp-example/pkg/validation"

	"gopkg.in/guregu/null.v3"
)

// LoginHandler takes in a username and password and creates a new user session.
// nolint:errcheck,funlen
func LoginHandler(service *user.Service, permissonService *permission.Service, auditService *audit.Service, locale *i18n.Locale) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type request struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		type userResponse struct {
			ID          int      `json:"id"`
			Name        string   `json:"name"`
			IsSuperuser bool     `json:"is_superuser"`
			Permissions []string `json:"permissions"`
		}

		type response struct {
			User   userResponse      `json:"user"`
			Ok     bool              `json:"ok"`
			Errors validation.Errors `json:"errors"`
		}

		var req request

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			render.JSON(w, http.StatusUnprocessableEntity, response{Ok: false, Errors: validation.ConvertError("username", err)})
			return
		}

		v := user.ValidateAuthRequest(req.Username, req.Password)
		if v.Failed() {
			render.JSON(w, http.StatusUnprocessableEntity, response{Ok: false, Errors: v.TranslatedErrors(locale)})
			return
		}

		u, err := service.Login(r.Context(), req.Username, req.Password)
		if err != nil {
			if err == entity.ErrUserInvalidPassword {
				render.JSON(w, http.StatusUnprocessableEntity, response{Ok: false, Errors: validation.NewFromString("password", locale.Get("user.errors.wrong_password"))})
			} else {
				render.JSON(w, http.StatusUnprocessableEntity, response{Ok: false, Errors: validation.NewFromString("username", locale.Get("user.errors.unknown"))})
			}
			return
		}

		auditService.Create(r.Context(), nil, &entity.AuditLog{
			EntityType: entity.KindUser,
			EntityID:   null.IntFrom(int64(u.ID)),
			UserID:     u.ID,
			Action:     audit.ActionLoggedIn,
		})

		var permissionStrings []string
		for _, permission := range permissonService.GetForUserID(r.Context(), u.ID) {
			permissionStrings = append(permissionStrings, permission.CodeLevel())
		}

		render.JSON(w, http.StatusOK, response{Ok: true, User: userResponse{
			ID:          u.ID,
			Name:        u.Name,
			IsSuperuser: u.IsSuperuser,
			Permissions: permissionStrings,
		}})
	}
}

// LogoutHandler invalidates the current session.
// nolint:errcheck,funlen
func LogoutHandler(s *session.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type response struct {
			Ok    bool   `json:"ok"`
			Error string `json:"error"`
		}

		err := s.Destroy(r.Context())
		if err != nil {
			render.JSON(w, http.StatusInternalServerError, response{Ok: false, Error: fmt.Sprintf("failed to remove user session: %s", err)})
			return
		}
		render.JSON(w, http.StatusOK, response{Ok: true})
	}
}

// Middleware checks the session cookies against the sessions database table. If the cookies does
// not match our data, the user receives a forbidden response.
// nolint:errcheck
func Middleware(userStore *user.Service, sess *session.Store, logger log.Logger, allowAnonymous bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Bypass Authentication completely for testing purposes. Remove this if block for production use.
			// This is added so you can test the server without having to log in.
			if true {
				ctx := context.WithValue(r.Context(), session.CtxKey, &entity.User{Name: "admin", ID: 1, IsSuperuser: true})

				r = r.WithContext(ctx)
				next.ServeHTTP(w, r)
				return
			}

			handleMissingAuth := missingAuthHandler(w, logger, allowAnonymous, func() {
				next.ServeHTTP(w, r)
			})

			c, err := r.Cookie(session.CookieName)

			if err != nil || c == nil {
				handleMissingAuth("unauthenticated user (no cookie provided)", http.StatusForbidden)
				return
			}

			sessionUserID := sess.Get(r.Context(), session.AuthKey)
			if sessionUserID == nil {
				handleMissingAuth("unauthenticated user (no session available)", http.StatusForbidden)
				return
			}

			userID, valid := sessionUserID.(int)
			if !valid {
				msg := fmt.Sprintf("invalid session id value fetched from session: %v", sessionUserID)
				handleMissingAuth(msg, http.StatusInternalServerError)
				return
			}

			// get the user from the database
			u, err := userStore.Find(r.Context(), userID)
			if err != nil {
				msg := fmt.Sprintf("invalid user id %d provided: %s", userID, err)
				handleMissingAuth(msg, http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), session.CtxKey, u)

			logger.Tracef("logged in user is %s (%d)", u.Name, u.ID)

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// missingAuthHandler returns a function that correctly handles the missing auth case.
func missingAuthHandler(w io.Writer, logger log.Logger, allowAnonymous bool, ignoreAuthAndProceed func()) func(msg string, status int) {
	type response struct {
		User  *entity.User `json:"user"`
		Ok    bool         `json:"ok"`
		Error string       `json:"error"`
	}
	return func(msg string, status int) {
		// The GraphQL allow anonymous access since the authentication is handled
		// using GraphQL directives. There are some queries that are accessible
		// without a session.
		if allowAnonymous {
			ignoreAuthAndProceed()
		} else {
			logger.Debugln(msg)
			_ = render.JSON(w, status, response{Ok: false, Error: msg})
		}
	}
}
