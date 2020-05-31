package endpoint

import (
	"context"

	kitendpoint "github.com/go-kit/kit/endpoint"

	"github.com/WiFeng/short-url/pkg/core/log"
	"github.com/WiFeng/short-url/pkg/service"
)

// Endpoints collects all of the endpoints that compose a profile service. It's
// meant to be used as a helper struct, to collect all of the endpoints into a
// single parameter.
type Endpoints struct {
	CreateEndpoint   kitendpoint.Endpoint
	QueryEndpoint    kitendpoint.Endpoint
	QueryAdvEndpoint kitendpoint.Endpoint
}

// New returns a Endpoints that wraps the provided server, and wires in all of the
// expected endpoint middlewares via the various parameters.
func New(s service.Service, logger log.Logger) Endpoints {
	var createEndpoint kitendpoint.Endpoint
	{
		createEndpoint = MakeCreateEndpoint(s)
		// createEndpoint = LoggingMiddleware(log.With(logger, "method", "Create"))(createEndpoint)
		createEndpoint = LoggingMiddleware(logger)(createEndpoint)
	}

	var queryEndpoint kitendpoint.Endpoint
	{
		queryEndpoint = MakeQueyrEndpoint(s)
		// queryEndpoint = LoggingMiddleware(log.With(logger, "method", "Query"))(queryEndpoint)
		queryEndpoint = LoggingMiddleware(logger)(queryEndpoint)
	}

	var queryAdvEndpoint kitendpoint.Endpoint
	{
		queryAdvEndpoint = MakeQueyrAdvEndpoint(s)
		// queryAdvEndpoint = LoggingMiddleware(log.With(logger, "method", "QueryAdv"))(queryAdvEndpoint)
		queryAdvEndpoint = LoggingMiddleware(logger)(queryAdvEndpoint)
	}

	return Endpoints{
		CreateEndpoint:   createEndpoint,
		QueryEndpoint:    queryEndpoint,
		QueryAdvEndpoint: queryAdvEndpoint,
	}
}

// MakeCreateEndpoint constructs a Create endpoint wrapping the service.
func MakeCreateEndpoint(s service.Service) kitendpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(CreateRequest)
		shortURL, err := s.Create(ctx, req.LongURL)
		return CreateResponse{ShortURL: shortURL, Err: err}, nil
	}
}

// MakeQueyrEndpoint constructs a Query endpoint wrapping the service.
func MakeQueyrEndpoint(s service.Service) kitendpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(QueryRequest)
		longURL, err := s.Query(ctx, req.ShortURL)
		return QueryResponse{LongURL: longURL, Err: err}, nil
	}
}

// MakeQueyrAdvEndpoint constructs a Query endpoint wrapping the service.
func MakeQueyrAdvEndpoint(s service.Service) kitendpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(QueryRequest)
		longURL, err := s.Query(ctx, req.ShortURL)
		return QueryResponse{LongURL: longURL, Err: err}, nil
	}
}

// compile time assertions for our response types implementing endpoint.Failer.
var (
	_ kitendpoint.Failer = CreateResponse{}
	_ kitendpoint.Failer = QueryResponse{}
)

// CreateRequest collects the request parameters for the Sum method.
type CreateRequest struct {
	LongURL string `json:"long_url"`
}

// CreateResponse collects the response values for the Sum method.
type CreateResponse struct {
	ShortURL string `json:"short_url"`
	Err      error  `json:"-"` // should be intercepted by Failed/errorEncoder
}

// Failed implements endpoint.Failer.
func (r CreateResponse) Failed() error { return r.Err }

// QueryRequest collects the request parameters for the Concat method.
type QueryRequest struct {
	ShortURL string `json:"short_url"`
}

// QueryResponse collects the response values for the Concat method.
type QueryResponse struct {
	LongURL string `json:"long_url"`
	Err     error  `json:"-"`
}

// Failed implements endpoint.Failer.
func (r QueryResponse) Failed() error { return r.Err }
