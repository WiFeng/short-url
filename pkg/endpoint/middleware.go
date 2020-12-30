package endpoint

import (
	"context"
	"time"

	kitendpoint "github.com/go-kit/kit/endpoint"

	"github.com/WiFeng/short-url/pkg/core/log"
)

// LoggingMiddleware returns an endpoint middleware that logs the
// duration of each invocation, and the resulting error, if any.
func LoggingMiddleware(logger log.Logger) kitendpoint.Middleware {
	return func(next kitendpoint.Endpoint) kitendpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			req, ok := request.(QueryRequest)
			if ok && req.ShortURL == "2bR" {
				logger = logger.With2("traceId", 1111)
			}

			defer func(begin time.Time) {
				// logger.Infow("defer caller", "transport_error", err, "took", time.Since(begin).Microseconds())
				log.Infow(ctx, "defer caller", "transport_error", err, "took", time.Since(begin).Microseconds())
			}(time.Now())
			return next(ctx, request)

		}
	}
}
