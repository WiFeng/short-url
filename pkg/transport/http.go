package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"

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
