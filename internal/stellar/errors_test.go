package stellar

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"testing"

	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/support/render/problem"
)

func TestHTTPStatus(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
		ok   bool
	}{
		{
			name: "wrapped response status",
			err:  errors.Join(errors.New("wrapper"), &horizonclient.Error{Response: &http.Response{StatusCode: http.StatusBadGateway}}),
			want: http.StatusBadGateway,
			ok:   true,
		},
		{
			name: "problem status",
			err:  &horizonclient.Error{Problem: problem.P{Status: http.StatusServiceUnavailable}},
			want: 503,
			ok:   true,
		},
		{
			name: "response status code",
			err:  &horizonclient.Error{Response: &http.Response{StatusCode: 404}},
			want: 404,
			ok:   true,
		},
		{
			name: "status line fallback",
			err:  &horizonclient.Error{Response: &http.Response{Status: "404 Not Found"}},
			want: 404,
			ok:   true,
		},
		{
			name: "non Horizon error",
			err:  errors.New("failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := HTTPStatus(tt.err)
			if got != tt.want || ok != tt.ok {
				t.Fatalf("HTTPStatus() = (%d, %t), want (%d, %t)", got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestIsNotFound(t *testing.T) {
	err := errors.New("not found")
	notFound := &horizonclient.Error{Response: &http.Response{StatusCode: http.StatusNotFound}}
	if !IsNotFound(errors.Join(err, notFound)) {
		t.Fatal("IsNotFound() = false for wrapped Horizon 404")
	}
	if IsNotFound(&horizonclient.Error{Response: &http.Response{StatusCode: http.StatusBadGateway}}) {
		t.Fatal("IsNotFound() = true for Horizon 502")
	}
}

func TestIsRetryableTransport(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "deadline", err: context.DeadlineExceeded, want: true},
		{name: "canceled", err: context.Canceled, want: true},
		{name: "EOF", err: io.EOF, want: true},
		{name: "unexpected EOF", err: io.ErrUnexpectedEOF, want: true},
		{name: "network timeout", err: &net.DNSError{IsTimeout: true}, want: true},
		{name: "application error", err: errors.New("invalid transaction")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRetryableTransport(tt.err); got != tt.want {
				t.Fatalf("IsRetryableTransport() = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestClassifySubmitError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want SubmitClass
	}{
		{
			name: "definite rejection",
			err:  &horizonclient.Error{Response: &http.Response{StatusCode: http.StatusBadRequest}},
			want: SubmitDefinite,
		},
		{
			name: "rate limited response",
			err:  &horizonclient.Error{Response: &http.Response{StatusCode: http.StatusTooManyRequests}},
			want: SubmitAmbiguous,
		},
		{
			name: "service error response",
			err:  &horizonclient.Error{Problem: problem.P{Status: http.StatusInternalServerError}},
			want: SubmitAmbiguous,
		},
		{name: "transport timeout", err: context.DeadlineExceeded, want: SubmitTransport},
		{name: "unknown error", err: errors.New("unexpected failure"), want: SubmitDefinite},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ClassifySubmitError(tt.err); got != tt.want {
				t.Fatalf("ClassifySubmitError() = %v, want %v", got, tt.want)
			}
		})
	}
}
