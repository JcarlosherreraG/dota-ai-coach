package errors

import (
	"errors"
	"fmt"
	"testing"
)

type mockNetError struct {
	isTimeout   bool
	isTemporary bool
}

func (e *mockNetError) Error() string   { return "net error" }
func (e *mockNetError) Timeout() bool   { return e.isTimeout }
func (e *mockNetError) Temporary() bool { return e.isTemporary }

func TestShouldRetry(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "No error (nil)",
			err:  nil,
			want: false,
		},
		{
			name: "Generic error (should fail)",
			err:  errors.New("some random error"),
			want: false,
		},
		{
			name: "API 200 OK",
			err:  NewAPIError(200, "ok"),
			want: false,
		},
		{
			name: "API 400 Bad Request (Fatal)",
			err:  NewAPIError(400, "bad request"),
			want: false,
		},
		{
			name: "API 401 Unauthorized (Fatal)",
			err:  NewAPIError(401, "auth failed"),
			want: false,
		},
		{
			name: "API 429 Too Many Requests (RETRY!)",
			err:  NewAPIError(429, "slow down"),
			want: true,
		},
		{
			name: "API 500 Internal Server Error (RETRY!)",
			err:  NewAPIError(500, "oops"),
			want: true,
		},
		{
			name: "API 503 Service Unavailable (RETRY!)",
			err:  NewAPIError(503, "down"),
			want: true,
		},
		{
			name: "API 599 (Border check)",
			err:  NewAPIError(599, "edge case"),
			want: true,
		},
		{
			name: "Net Timeout (RETRY!)",
			err:  &mockNetError{isTimeout: true},
			want: true,
		},
		{
			name: "Net Temporary (RETRY!)",
			err:  &mockNetError{isTemporary: true},
			want: true,
		},
		{
			name: "Net Fatal (DNS error etc)",
			err:  &mockNetError{isTimeout: false, isTemporary: false},
			want: false,
		},
		{
			name: "Wrapped API Error 502",
			err:  fmt.Errorf("context: %w", NewAPIError(502, "wrapped")),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShouldRetry(tt.err); got != tt.want {
				t.Errorf("ShouldRetry() = %v, want %v (err: %v)", got, tt.want, tt.err)
			}
		})
	}
}
