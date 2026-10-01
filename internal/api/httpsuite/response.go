package httpsuite

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

// Response represents the structure of an HTTP response, including a status code, message, and optional body.
type Response[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Body    T      `json:"body"`
}

type emptyResponse struct{}

func NewEmptyResponse() *emptyResponse {
	return &emptyResponse{}
}

// Marshal serializes the Response struct into a JSON byte slice.
// It logs an error if marshalling fails.
func (r *Response[T]) Marshal() ([]byte, error) {
	jsonResponse, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return jsonResponse, nil
}

// SendResponse creates a Response struct, serializes it to JSON, and writes it to the provided http.ResponseWriter.
// If the body parameter is non-nil, it will be included in the response body.
func SendResponse[T any](ctx context.Context, w http.ResponseWriter, message string, code int, body *T) {
	response := &Response[T]{
		Code:    code,
		Message: message,
	}
	if body != nil {
		response.Body = *body
	}

	writeJSONResponse(ctx, w, response)
}

// SendEmptyResponse creates a Response struct with an empty body, serializes it to JSON, and writes it to the provided http.ResponseWriter.
func SendEmptyResponse(ctx context.Context, w http.ResponseWriter, message string, code int) {
	response := &Response[emptyResponse]{
		Code:    code,
		Message: message,
		Body:    *NewEmptyResponse(),
	}

	writeJSONResponse(ctx, w, response)
}

// writeResponse serializes a Response and writes it to the http.ResponseWriter with appropriate headers.
// If an error occurs during the write, it logs the error and sends a 500 Internal Server Error response.
func writeJSONResponse[T any](ctx context.Context, w http.ResponseWriter, r *Response[T]) {
	jsonResponse, err := r.Marshal()
	if err != nil {
		slog.ErrorContext(ctx, "error marshalling response", "error", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.Code)

	_, err = w.Write(jsonResponse)
	if err != nil {
		slog.ErrorContext(ctx, "error writing response", "error", err)
	}
}

type (
	ResponseConfig struct {
		StrictMode bool
	}

	ResponseOption func(*ResponseConfig)
)

func newResponseConfig(opts ...ResponseOption) *ResponseConfig {
	config := &ResponseConfig{
		StrictMode: false,
	}

	for _, opt := range opts {
		opt(config)
	}

	return config
}

// WithStrictMode is a ResponseOption that sets DisallowedUnknownFields on the JSON decoder,
// which will cause an error if the request body contains fields that are not present in the output structure.
func WithStrictMode(strict bool) ResponseOption {
	return func(config *ResponseConfig) {
		config.StrictMode = strict
	}
}

// ReadResponse reads the body of an HTTP request and decodes it into the provided output structure.
// It closes the request body after reading and logs any errors that occur during the process.
func ReadResponse(w http.ResponseWriter, r *http.Request, out any, opts ...ResponseOption) error {
	options := newResponseConfig(opts...)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			slog.ErrorContext(r.Context(), "error closing response body", "error", err)
		}
	}()

	// Limit the size of the request body to 10MB
	closer := http.MaxBytesReader(w, r.Body, 10*1024*1024)

	decoder := json.NewDecoder(closer)
	if options.StrictMode {
		decoder.DisallowUnknownFields()
	}
	if err := decoder.Decode(out); err != nil {
		return fmt.Errorf("error decoding response body: %w", err)
	}

	return nil
}
