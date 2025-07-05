package auth

import (
	"context"

	"github.com/QuangTung97/weblib/auth/oauth"
	"github.com/QuangTung97/weblib/dblib"
	"github.com/QuangTung97/weblib/router"
)

type Service[User any] interface {
	NewMiddleware() router.Middleware

	HandleLogin(ctx router.Context, info oauth.GoogleAccount) error

	GetUser(ctx context.Context) User
	GetUserOptional(ctx context.Context) (User, bool)
}

type UserSession struct {
}

type Repository[User any] interface {
	InsertUser(ctx context.Context, user *User) error
	InsertUserSession(ctx context.Context, user User, sessionID string) error
	GetUserBySession(ctx context.Context, sessionID string) (User, error)
}

type serviceImpl[User any] struct {
}

func NewService[User any](
	provider dblib.Provider,
	repo Repository[User],
) Service[User] {
	return &serviceImpl[User]{}
}

func (s *serviceImpl[User]) NewMiddleware() router.Middleware {
	return nil
}

func (s *serviceImpl[User]) HandleLogin(ctx router.Context, info oauth.GoogleAccount) error {
	return nil
}

func (s *serviceImpl[User]) GetUser(ctx context.Context) User {
	var empty User
	return empty
}

func (s *serviceImpl[User]) GetUserOptional(ctx context.Context) (User, bool) {
	var empty User
	return empty, false
}
