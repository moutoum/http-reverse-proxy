package cache

import "net/http"

// Request is a wrapper around a http.Request.
// It helps to deal with the cache mechanisms.
type Request struct {

	// request is the wrapped http.Request.
	request *http.Request

	// cacheControl is the parsed "Cache-Control" http header
	// representation for the wrapped request.
	cacheControl *CacheControl

	// key is an unique identifier for the wrapped request.
	key string
}

// NewRequest creates a Request from a http.Request.
func NewRequest(r *http.Request) *Request {
	return &Request{
		request:      r,
		cacheControl: ParseCacheControl(r.Header.Get("Cache-Control")),
		key:          generateRequestKey(r),
	}
}

// IsCacheable checks if the wrapped request is able to be
// cached.
// To be cacheable, the request has to have the good HTTP method (GET, HEAD)
// and having a max-age greater than zero.
func (r *Request) IsCacheable() bool {
	if r.request.Method != http.MethodGet && r.request.Method != http.MethodHead {
		return false
	}

	// TODO: Add strong and smooth validations (ETag, If-Match...).

	if r.cacheControl.HasMaxAge && (r.cacheControl.MaxAge == 0) {
		return false
	}

	if !r.cacheControl.Public {
		return false
	}

	return true
}

// generateRequestKey generates an unique identifier
// from a http.Request.
func generateRequestKey(r *http.Request) string {
	return r.URL.RequestURI()
}
