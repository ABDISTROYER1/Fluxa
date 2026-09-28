package idempotency

import (
	"bytes"
	"errors"
	"net/http"
)

var errResponseTooLarge = errors.New("idempotent response exceeds configured size limit")

// recorder captures a bounded response without exposing it to the client.
// The real ResponseWriter is not touched until Complete has committed.
type recorder struct {
	header      http.Header
	body        bytes.Buffer
	status      int
	wroteHeader bool
	maxBodySize int
	err         error
}

func newRecorder(maxBodySize int) *recorder {
	return &recorder{
		header:      make(http.Header),
		status:      http.StatusOK,
		maxBodySize: maxBodySize,
	}
}

func (r *recorder) Header() http.Header {
	return r.header
}

func (r *recorder) WriteHeader(status int) {
	if r.wroteHeader {
		return
	}
	r.status = status
	r.wroteHeader = true
}

func (r *recorder) Write(b []byte) (int, error) {
	if r.err != nil {
		return 0, r.err
	}
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	remaining := r.maxBodySize - r.body.Len()
	if remaining < 0 {
		remaining = 0
	}
	if len(b) > remaining {
		if remaining > 0 {
			_, _ = r.body.Write(b[:remaining])
		}
		r.err = errResponseTooLarge
		return 0, r.err
	}
	return r.body.Write(b)
}

func (r *recorder) response() Response {
	return Response{
		Status:  r.status,
		Headers: r.header.Clone(),
		Body:    append([]byte(nil), r.body.Bytes()...),
	}
}
