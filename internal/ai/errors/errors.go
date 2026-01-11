package errors

import (
	"errors"
	"fmt"
	"net"
)

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("api error (status %d): %s", e.StatusCode, e.Message)
}

func NewAPIError(code int, msg string) error {
	return &APIError{StatusCode: code, Message: msg}
}

func ShouldRetry(err error) bool {
	if err == nil {
		return false
	}

	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 429 || (apiErr.StatusCode >= 500 && apiErr.StatusCode < 600)
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout() || netErr.Temporary()
	}

	return false
}
