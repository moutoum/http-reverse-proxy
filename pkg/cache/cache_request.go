package cache

import (
	"net/http"
	"time"
)

type Request struct {
	request      *http.Request
	CacheControl *CacheControl
	time         time.Time
	key          string
}

func NewRequest(r *http.Request) *Request {
	return &Request{
		request:      r,
		CacheControl: ParseCacheControl(r.Header.Get("Cache-Control")),
		time:         time.Now(),
		key:          generateRequestKey(r),
	}
}

func (r *Request) IsCacheable() bool {
	if r.request.Method != http.MethodGet && r.request.Method != http.MethodHead {
		return false
	}

	// TODO: Add strong and smooth validations (ETag, If-Match...).

	if r.CacheControl.HasMaxAge && (r.CacheControl.MaxAge == 0) {
		return false
	}

	return true
}

func generateRequestKey(r *http.Request) string {
	return r.RequestURI
}
