package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	jaegerclient "github.com/uber/jaeger-client-go"

	kitendpoint "github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"

	"github.com/WiFeng/short-url/pkg/core/log"
	"github.com/WiFeng/short-url/pkg/endpoint"
)

var (
	// ErrBadRouting is returned when an expected path variable is missing.
	// It always indicates programmer error.
	ErrBadRouting = errors.New("inconsistent mapping between route and handler (programmer error)")

	// ErrReponseAssert is the errof of type asserting
	ErrReponseAssert = errors.New("response assert error")
)

// NewHTTPHandler returns an HTTP handler that makes a set of endpoints
// available on predefined paths.
func NewHTTPHandler(endpoints endpoint.Endpoints, logger log.Logger) http.Handler {
	r := mux.NewRouter()
	options := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(errorEncoder),
		kithttp.ServerErrorLogger(logger),
		kithttp.ServerBefore(beforeHandler),
		kithttp.ServerAfter(afterHandler),
	}

	r.Methods("POST").Path("/admin/create").Handler(kithttp.NewServer(
		endpoints.CreateEndpoint,
		decodeHTTPCreateRequest,
		encodeHTTPGenericResponse,
		options...,
	))

	r.Methods("POST").Path("/admin/query").Handler(kithttp.NewServer(
		endpoints.QueryEndpoint,
		decodeHTTPQueryRequest,
		encodeHTTPGenericResponse,
		options...,
	))

	r.Methods("GET").Path("/x/{id}").Handler(kithttp.NewServer(
		endpoints.QueryAdvEndpoint,
		decodeHTTPQueryAdvRequest,
		encodeHTTPQueryAdvResponse,
		options...,
	))

	return r
}

func errorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	w.WriteHeader(err2code(err))
	json.NewEncoder(w).Encode(errorWrapper{Error: err.Error()})
}

func err2code(err error) int {
	//switch err {
	//case service.ErrTwoZeroes, service.ErrMaxSizeExceeded, service.ErrIntOverflow:
	//	return http.StatusBadRequest
	//}
	return http.StatusInternalServerError
}

func errorDecoder(r *http.Response) error {
	var w errorWrapper
	if err := json.NewDecoder(r.Body).Decode(&w); err != nil {
		return err
	}
	return errors.New(w.Error)
}

type errorWrapper struct {
	Error string `json:"error"`
}

func startSpan(ctx context.Context, r *http.Request) context.Context {
	var serverSpan opentracing.Span
	var appSpecificOperationName = fmt.Sprintf("[%s]%s", r.Method, r.URL)

	wireContext, err := opentracing.GlobalTracer().Extract(
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(r.Header))
	if err != nil {
		// Optionally record something about err here
	}

	// Create the span referring to the RPC client if available.
	// If wireContext == nil, a root span will be created.
	serverSpan = opentracing.StartSpan(
		appSpecificOperationName,
		ext.RPCServerOption(wireContext))

	// defer serverSpan.Finish()

	// ctx = opentracing.ContextWithSpan(context.Background(), serverSpan)
	newCtx := opentracing.ContextWithSpan(ctx, serverSpan)

	return newCtx
}

func finishSpan(ctx context.Context, w http.ResponseWriter) context.Context {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		span.Finish()
	}
	return ctx
}

func getTraceID(ctx context.Context) string {
	var traceID string
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		spanContext := span.Context()
		jeagerSpanContext, ok := spanContext.(jaegerclient.SpanContext)
		if ok {
			traceID = jeagerSpanContext.TraceID().String()
		}
	}

	return traceID
}

func buildLogger(ctx context.Context) context.Context {
	newLogg := log.GetDefaultLogger()

	traceID := getTraceID(ctx)
	if traceID != "" {
		newLogg = newLogg.With2("traceid", traceID)
	} else {
		newLogg = newLogg.With2("traceid", traceID)
	}

	newCtx := log.ContextWithLogger(ctx, newLogg)
	return newCtx
}

func beforeHandler(ctx context.Context, r *http.Request) context.Context {
	ctx = startSpan(ctx, r)
	ctx = buildLogger(ctx)
	return ctx
}

func afterHandler(ctx context.Context, w http.ResponseWriter) context.Context {
	ctx = finishSpan(ctx, w)
	return ctx
}

// decodeHTTPSumRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded sum request from the HTTP request body. Primarily useful in a
// server.
func decodeHTTPCreateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoint.CreateRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

func decodeHTTPQueryRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoint.QueryRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

func decodeHTTPQueryAdvRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoint.QueryRequest
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrBadRouting
	}
	req.ShortURL = id
	return req, nil
}

// encodeHTTPGenericRequest is a transport/http.EncodeRequestFunc that
// JSON-encodes any request to the request body. Primarily useful in a client.
func encodeHTTPGenericRequest(_ context.Context, r *http.Request, request interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(request); err != nil {
		return err
	}
	r.Body = ioutil.NopCloser(&buf)
	return nil
}

// encodeHTTPGenericResponse is a transport/http.EncodeResponseFunc that encodes
// the response as JSON to the response writer. Primarily useful in a server.
func encodeHTTPGenericResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if f, ok := response.(kitendpoint.Failer); ok && f.Failed() != nil {
		errorEncoder(ctx, f.Failed(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

// encodeHTTPQueryAdvResponse
func encodeHTTPQueryAdvResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if f, ok := response.(kitendpoint.Failer); ok && f.Failed() != nil {
		errorEncoder(ctx, f.Failed(), w)
		return nil
	}
	resp, ok := response.(endpoint.QueryResponse)
	if !ok {
		errorEncoder(ctx, ErrReponseAssert, w)
		return nil
	}

	w.Header().Set("Location", resp.LongURL)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusFound)

	return nil
}
