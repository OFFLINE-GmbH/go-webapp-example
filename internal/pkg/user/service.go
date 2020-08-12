package user

import (
	"context"

	"go-webapp-example/internal/pkg/entity"
	"go-webapp-example/pkg/session"

	"golang.org/x/crypto/bcrypt"

	"github.com/pkg/errors"
)

// Service is used to interact with the entity. It
// allows access to the store by embedding it.
type Service struct {
	*Store
	Session sessionHandler
}

// sessionHandler defines the needed methods to store user sessions.
type sessionHandler interface {
	RenewToken(ctx context.Context) error
	Put(ctx context.Context, key string, val interface{})
}

// NewService returns a pointer to a new Service.
func NewService(store *Store, sess sessionHandler) *Service {
	return &Service{
		Store:   store,
		Session: sess,
	}
}

// Create creates a new user account. The password is hashed automatically.
func (s Service) Create(ctx context.Context, u *entity.User) (*entity.User, error) {
	u, err := hashPassword(u)
	if err != nil {
		return u, errors.WithStack(err)
	}

	return s.Store.Create(ctx, u)
}

// Update updates a user account. The password is hashed automatically.
func (s Service) Update(ctx context.Context, u *entity.User) (*entity.User, error) {
	if u.Password != "" {
		hashed, err := hashPassword(u)
		if err != nil {
			return hashed, errors.WithStack(err)
		}
	}

	return s.Store.Update(ctx, u)
}

// hashPassword hashes and sets the user's password.
func hashPassword(u *entity.User) (*entity.User, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.Wrap(err, "failed to hash user password")
	}
	u.Password = string(bytes)
	return u, nil
}

// Login checks if a user with given credentials exists in the database and creates a session.
func (s Service) Login(ctx context.Context, username, password string) (*entity.User, error) {
	user, err := s.Store.FindByName(ctx, username)
	if err != nil {
		return user, entity.ErrNotFound
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, entity.ErrUserInvalidPassword
	}

	// First renew the session token.
	err = s.Session.RenewToken(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to renew session token")
	}

	// Then make the privilege-level change.
	s.Session.Put(ctx, session.AuthKey, user.ID)

	return user, nil
}
