package auth

import (
	"context"

	"github.com/QuangTung97/weblib/router"
)

type Service[User any] interface {
	NewMiddleware() router.Middleware

	GetUser(ctx context.Context) User
	GetUserOptional(ctx context.Context) (User, bool)
}
