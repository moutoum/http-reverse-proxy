package integration

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/moutoum/http-reverse-proxy/pkg/proxy"
	"github.com/stretchr/testify/assert"
)

func TestProxy(t *testing.T) {
	t.Run("Working scenarios", func(t *testing.T) {
		originServer := NewOriginServer().
			WithRouteStatus("/api/check", http.StatusOK).
			WithRouteContent("/api/data", http.StatusOK, []byte("My super content")).
			WithRouteStatus("/api/not-found", http.StatusNotFound).
			Start()

		defer originServer.Close()

		proxyServer := proxy.New(originServer.URL())

		t.Run("Status code forwarding", func(t *testing.T) {
			assert.HTTPSuccess(t, proxyServer.ServeHTTP, "GET", "/api/check", nil)
			assert.HTTPSuccess(t, proxyServer.ServeHTTP, "GET", "/api/data", nil)
			assert.HTTPStatusCode(t, proxyServer.ServeHTTP, "GET", "/api/not-found", nil, http.StatusNotFound)
		})

		t.Run("Body forwarding", func(t *testing.T) {
			assert.HTTPBodyContains(t, proxyServer.ServeHTTP, "GET", "/api/data", nil, "My super content")
		})
	})

	t.Run("Invalid origin server", func(t *testing.T) {
		invalidURL := &url.URL{}
		proxyServer := proxy.New(invalidURL)
		assert.HTTPStatusCode(t, proxyServer.ServeHTTP, "GET", "/api/data", nil, http.StatusBadGateway)
	})
}
