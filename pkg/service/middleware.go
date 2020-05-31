package service

import (
	"context"

	"github.com/WiFeng/short-url/pkg/core/log"
)

// Middleware describes a service (as opposed to endpoint) middleware.
type Middleware func(Service) Service

// LoggingMiddleware takes a logger as a dependency
// and returns a ServiceMiddleware.
func LoggingMiddleware(logger log.Logger) Middleware {
	return func(next Service) Service {
		return loggingMiddleware{logger, next}
	}
}

type loggingMiddleware struct {
	logger log.Logger
	next   Service
}

func (mw loggingMiddleware) Create(ctx context.Context, longURL string) (shortURL string, err error) {
	defer func() {
		mw.logger.Infow("defer caller", "method", "Create", "longURL", longURL, "shortURL", shortURL, "err", err)
	}()
	return mw.next.Create(ctx, longURL)
}

func (mw loggingMiddleware) Query(ctx context.Context, shortURL string) (longURL string, err error) {
	defer func() {
		mw.logger.Infow("defer caller", "method", "Query", "shortURL", shortURL, "longURL", longURL, "err", err)
	}()
	return mw.next.Query(ctx, shortURL)
}
