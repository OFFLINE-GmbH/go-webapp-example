package session

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"go-webapp-example/internal/pkg/entity"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
)

// CookieName is the name of the session cookie.
const CookieName = "gowebapp_session"

// CtxKey is used to derive the current user from a context.
var CtxKey = &contextKey{"user"}

// contextKey is the struct that wraps the currently logged in user in the context.
type contextKey struct {
	name string
}

// UserFromContext tries to fetch the currently logged in user from the request
// context and returns it if available.
func UserFromContext(ctx context.Context) (*entity.User, error) {
	user, ok := ctx.Value(CtxKey).(*entity.User)
	if !ok {
		return nil, fmt.Errorf("invalid auth user received from context")
	}
	return user, nil
}

// AuthKey is the session key that contains the current user's id.
var AuthKey = "user_id"

type Store struct {
	session *scs.SessionManager
}

func New(db *sql.DB) *Store {
	sess := scs.New()
	sess.Lifetime = 24 * 7 * time.Hour
	sess.IdleTimeout = time.Hour
	sess.Cookie.Name = CookieName
	sess.Cookie.HttpOnly = true
	sess.Cookie.Persist = true
	sess.Cookie.SameSite = http.SameSiteStrictMode
	sess.Cookie.Secure = false

	sess.Store = mysqlstore.NewWithCleanupInterval(db, time.Hour)

	return &Store{
		session: sess,
	}
}

func (s *Store) Put(ctx context.Context, key string, val interface{}) {
	s.session.Put(ctx, key, val)
}

func (s *Store) Get(ctx context.Context, key string) interface{} {
	return s.session.Get(ctx, key)
}

func (s *Store) RenewToken(ctx context.Context) error {
	return s.session.RenewToken(ctx)
}

func (s *Store) Destroy(ctx context.Context) error {
	return s.session.Destroy(ctx)
}

// LoadAndSave provides middleware which automatically loads and saves session
// data for the current request, and communicates the session token to and from
// the client in a cookie.
func (s *Store) Middleware(next http.Handler) http.Handler {
	return s.session.LoadAndSave(next)
}
