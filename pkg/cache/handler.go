package cache

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

var cacheableStatus = map[int]bool{
	http.StatusOK:               true,
	http.StatusFound:            true,
	http.StatusNotFound:         true,
	http.StatusNotModified:      true,
	http.StatusMovedPermanently: true,
}

type Handler struct {
	Cache           Cache
	Origin          http.Handler
	cacheableStatus map[int]bool
}

func NewHandler(c Cache, o http.Handler) *Handler {
	h := &Handler{
		Cache:           c,
		Origin:          o,
		cacheableStatus: make(map[int]bool, len(cacheableStatus)),
	}

	for k, v := range cacheableStatus {
		h.cacheableStatus[k] = v
	}

	return h
}

func (h *Handler) ServeHTTP(writer http.ResponseWriter, r *http.Request) {
	request := NewRequest(r)

	if !request.IsCacheable() {
		logrus.Debug("Not cacheable")
		h.forwardToOrigin(writer, request)
		return
	}

	resource := h.load(request)
	if resource != nil {
		logrus.Debug("Forwarding resource to client")
		h.forwardResourceToClient(resource, writer)
		return
	}

	logrus.Debug("No resources matched")

	if request.CacheControl.OnlyCached {
		writer.WriteHeader(http.StatusGatewayTimeout)
		return
	}

	rw := NewResourceWriter(writer)
	h.forwardToOrigin(rw, request)
	resource = rw.Resource()

	if cacheable := h.isResourceCacheable(resource); !cacheable {
		return
	}

	logrus.Debug("Storing resource in cache")
	h.Cache.Store(request.key, resource)
}

func (h *Handler) forwardToOrigin(writer http.ResponseWriter, request *Request) {
	h.Origin.ServeHTTP(writer, request.request)
}

func (h *Handler) load(request *Request) *Resource {
	return h.Cache.Get(request.key)
}

func (h *Handler) forwardResourceToClient(r *Resource, writer http.ResponseWriter) {
	for header, values := range r.Headers {
		for _, value := range values {
			writer.Header().Add(header, value)
		}
	}

	writer.WriteHeader(r.Status)
	_, _ = writer.Write(r.Body)
}

func (h *Handler) isResourceCacheable(resource *Resource) bool {
	cc := ParseCacheControl(resource.Headers.Get("Cache-Control"))

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