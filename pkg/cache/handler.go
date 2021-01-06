package cache

import (
	"math"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"
)

// cacheableStatus is the default list of the
// available response status to be stored in cache.
var cacheableStatus = map[int]bool{
	http.StatusOK:               true,
	http.StatusFound:            true,
	http.StatusNotFound:         true,
	http.StatusNotModified:      true,
	http.StatusMovedPermanently: true,
}

// Handler is a http.Handler that is used as a middleware
// to enable a caching feature.
type Handler struct {

	// Cache is the cache storage behavior to use to store
	// the http responses.
	Cache Cache

	// Origin is the http handler that will have the caching feature
	// in front of it.
	Origin http.Handler

	// cacheableStatus is at first a copy of the global variable. It will
	// help to have customized response status for the current cache
	// instance.
	cacheableStatus map[int]bool
}

// NewHandler creates a cache middle instance from a cache storage
// behavior and a http.Handler.
func NewHandler(c Cache, o http.Handler) *Handler {
	h := &Handler{
		Cache:           c,
		Origin:          o,
		cacheableStatus: make(map[int]bool, len(cacheableStatus)),
	}

	// Deep copy of the cacheableStatus global variable.
	for k, v := range cacheableStatus {
		h.cacheableStatus[k] = v
	}

	return h
}

// ServeHTTP adds the cache behavior in front of the origin handler.
//
// It receives all the requests and internally checks the status of the
// requested resource before deciding to ask the origin server, or to use
// the internal cached response.
// The client and the origin server can control the cache behavior with
// the "Cache-Control" http header.
//
// More details can be found here: https://tools.ietf.org/html/rfc7234
func (h *Handler) ServeHTTP(writer http.ResponseWriter, r *http.Request) {
	request := NewRequest(r)

	// If the incoming request is not cacheable for some reasons,
	// we directly forward the request to the origin server.
	if !request.IsCacheable() {
		logrus.Debug("Not cacheable")
		h.Origin.ServeHTTP(writer, request.request)
		return
	}

	// Try loading a cached resource.
	resource := h.load(request)
	if resource != nil {
		if resource.cc.HasMaxAge {
			// If the resource is found, it checks if it's stale
			// or valid, and ask the origin server again if needed.
			acceptedMaxAge := resource.cc.MaxAge

			// Max stale is a client option that allows the resource to
			// be stale but since some specified time.
			if request.cacheControl.HasMaxStale {
				acceptedMaxAge += request.cacheControl.MaxStale
			}

			// Min fresh is a client option that force the resource to
			// be fresh for at least some time after the request.
			if request.cacheControl.HasMinFresh {
				acceptedMaxAge -= request.cacheControl.MinFresh
			}

			if resource.Age() < acceptedMaxAge {
				logrus.Debug("Forwarding resource to client")
				forwardResource(resource, writer)
				return
			}

			// TODO: Smooth validation and update resource freshness (304, ETag...).

		} else {
			logrus.Debug("Forwarding resource to client")
			forwardResource(resource, writer)
			return
		}
	}

	logrus.WithField("resource", request.request.URL.RequestURI()).Debug("No resources matched in cache")

	// Only cached is a client option that force the request to use the
	// cached response. So, if the resource is not available, we send back
	// an http error (502) to the client.
	if request.cacheControl.OnlyCached {
		writer.WriteHeader(http.StatusGatewayTimeout)
		return
	}

	h.forwardToOrigin(writer, request)
}

// forwardToOrigin sends the request to the origin server and try
// to save the response in the cache.
func (h *Handler) forwardToOrigin(writer http.ResponseWriter, request *Request) {
	rw := NewResourceWriter()
	h.Origin.ServeHTTP(rw, request.request)
	resource := rw.Resource()
	forwardResource(resource, writer)

	if !h.isResourceCacheable(resource) {
		return
	}

	logrus.WithField("resource", request.request.URL.RequestURI()).Debug("Storing resource in cache")
	h.Cache.Store(request.key, resource)
}

// load gets the resource from the cache that matches the
// given request.
func (h *Handler) load(request *Request) *Resource {
	return h.Cache.Get(request.key)
}

// isResourceCacheable checks if the given response resource can be
// saved in cache or not. To be valid, the resource has to have a valid
// http status, and having coherent Cache-Control options.
func (h *Handler) isResourceCacheable(resource *Resource) bool {
	cc := resource.cc

	if cacheable, ok := h.cacheableStatus[resource.Status]; !ok || !cacheable {
		return false
	}

	if cc.NoStore || cc.NoCache {
		return false
	}

	if !cc.Public {
		return false
	}

	if cc.HasMaxAge && cc.MaxAge == 0 {
		return false
	}

	return true
}

// forwardResource pipe the given resource to the given writer.
// It adds the cache HTTP headers if needed (e.g "Age").
func forwardResource(r *Resource, writer http.ResponseWriter) {
	if r.cc.HasMaxAge {
		age := r.Age().Seconds()
		age = math.Floor(age)
		writer.Header().Set("Age", strconv.Itoa(int(age)))
	}

	for header, values := range r.Headers {
		for _, value := range values {
			writer.Header().Add(header, value)
		}
	}

	writer.WriteHeader(r.Status)
	_, _ = writer.Write(r.Body)
}
