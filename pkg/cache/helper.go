package cache

import (
	"net/http"
	"time"
)

// ResourceWriter is a Resource generator that implements
// http.ResponseWriter. It is used as a buffer to store
// the origin server response and generate a resource from it.
type ResourceWriter struct {
	status  int
	headers http.Header
	body    []byte
}

// NewResourceWriter creates an empty ResourceWriter.
func NewResourceWriter() *ResourceWriter {
	return &ResourceWriter{
		headers: make(http.Header),
	}
}

// Header is the "http.ResponseWriter" interface implementation.
func (r *ResourceWriter) Header() http.Header {
	return r.headers
}

// Write is the "http.ResponseWriter" interface implementation.
func (r *ResourceWriter) Write(bytes []byte) (int, error) {
	r.body = bytes

	if r.status == 0 {
		r.status = http.StatusOK
	}

	return len(bytes), nil
}

// WriteHeader is the "http.ResponseWriter" interface implementation.
func (r *ResourceWriter) WriteHeader(statusCode int) {
	r.status = statusCode
}

// Resource generates a Resource from its internal state.
// NOTE: This function has to be used only when the origin
//       server finished to write the response into.
func (r *ResourceWriter) Resource() *Resource {
	return &Resource{
		Status:  r.status,
		Headers: r.headers,
		Body:    r.body,
		Date:    time.Now(),
		cc:      ParseCacheControl(r.headers.Get("Cache-Control")),
	}
}
