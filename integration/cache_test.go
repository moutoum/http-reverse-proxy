package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/moutoum/http-reverse-proxy/pkg/cache"
	"github.com/stretchr/testify/assert"
)

func TestCache_SimpleStatus(t *testing.T) {
	origin := NewMockHandler()
	origin.StatusRoute("/api/ok", http.StatusOK, nil).Once()
	origin.StatusRoute("/api/not-found", http.StatusNotFound, nil).Once()
	c := cache.NewHandler(cache.NewInMemoryCache(), origin)

	assert.HTTPSuccess(t, c.ServeHTTP, "GET", "/api/ok", nil)
	assert.HTTPSuccess(t, c.ServeHTTP, "GET", "/api/ok", nil)
	assert.HTTPStatusCode(t, c.ServeHTTP, "GET", "/api/not-found", nil, http.StatusNotFound)
	assert.HTTPStatusCode(t, c.ServeHTTP, "GET", "/api/not-found", nil, http.StatusNotFound)

	origin.AssertExpectations(t)
}

func TestCache_SimpleStatusAndBody(t *testing.T) {
	origin := NewMockHandler()
	origin.DataRoute("/api/data", http.StatusOK, []byte("My body content"), nil).Once()
	c := cache.NewHandler(cache.NewInMemoryCache(), origin)

	// First call that fetches the origin server.
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/data", nil)
	c.ServeHTTP(recorder, req)
	assert.Equal(t, recorder.Code, http.StatusOK)
	assert.Equal(t, recorder.Body.String(), "My body content")

	// Second call that uses the cache.
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/data", nil)
	c.ServeHTTP(recorder, req)
	assert.Equal(t, recorder.Code, http.StatusOK)
	assert.Equal(t, recorder.Body.String(), "My body content")

	origin.AssertExpectations(t)
}

func TestCache_NonCacheableStatus(t *testing.T) {
	origin := NewMockHandler()
	origin.StatusRoute("/api/bad-gateway", http.StatusBadGateway, nil).Twice()
	c := cache.NewHandler(cache.NewInMemoryCache(), origin)

	assert.HTTPStatusCode(t, c.ServeHTTP, "GET", "/api/bad-gateway", nil, http.StatusBadGateway)
	assert.HTTPStatusCode(t, c.ServeHTTP, "GET", "/api/bad-gateway", nil, http.StatusBadGateway)

	origin.AssertExpectations(t)
}

func TestCache_NonCacheableMethod(t *testing.T) {
	origin := NewMockHandler()
	origin.StatusRoute("/api/post-request", http.StatusOK, nil).Twice()
	c := cache.NewHandler(cache.NewInMemoryCache(), origin)

	assert.HTTPSuccess(t, c.ServeHTTP, "POST", "/api/post-request", nil)
	assert.HTTPSuccess(t, c.ServeHTTP, "POST", "/api/post-request", nil)

	origin.AssertExpectations(t)
}

func TestCache_ServerControl(t *testing.T) {
	t.Parallel()

	origin := NewMockHandler()
	c := cache.NewHandler(cache.NewInMemoryCache(), origin)

	t.Run("Max age equal 0 (No cache)", func(t *testing.T) {
		origin.StatusRoute("/api/max-age/0", http.StatusOK, map[string]string{"Cache-Control": "max-age=0"}).Twice()
		assert.HTTPSuccess(t, c.ServeHTTP, "GET", "/api/max-age/0", nil)
		assert.HTTPSuccess(t, c.ServeHTTP, "GET", "/api/max-age/0", nil)
	})

	t.Run("Max age equal 3 (3s cache)", func(t *testing.T) {
		t.Parallel()
		origin.StatusRoute("/api/max-age/3", http.StatusOK, map[string]string{"Cache-Control": "max-age=3"}).Twice()

		// First call that uses origin server.
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/max-age/3", nil)
		c.ServeHTTP(recorder, req)
		assert.Equal(t, recorder.Code, http.StatusOK)
		// Checking the headers.
		maxAgeH, ageH := recorder.Header().Get("Cache-Control"), recorder.Header().Get("Age")
		assert.Equal(t, "max-age=3", maxAgeH)
		assert.Equal(t, "0", ageH)

		time.Sleep(1 * time.Second)
		// Second call that uses cache.
		recorder = httptest.NewRecorder()
		req = httptest.NewRequest("GET", "/api/max-age/3", nil)
		c.ServeHTTP(recorder, req)
		assert.Equal(t, recorder.Code, http.StatusOK)
		// Checking header age progression.
		maxAgeH, ageH = recorder.Header().Get("Cache-Control"), recorder.Header().Get("Age")
		assert.Equal(t, "max-age=3", maxAgeH)
		assert.Equal(t, "1", ageH)

		// Waiting for the resource to expire. It should call the origin server next time.
		time.Sleep(2 * time.Second)
		assert.HTTPSuccess(t, c.ServeHTTP, "GET", "/api/max-age/3", nil)
	})

	t.Run("No-store option enabled", func(t *testing.T) {
		origin.StatusRoute("/api/no-store", http.StatusOK, map[string]string{"Cache-Control": "no-store"}).Twice()
		assert.HTTPSuccess(t, c.ServeHTTP, "GET", "/api/no-store", nil)
		assert.HTTPSuccess(t, c.ServeHTTP, "GET", "/api/no-store", nil)
	})

	t.Run("No-cache option enabled", func(t *testing.T) {
		origin.StatusRoute("/api/no-cache", http.StatusOK, map[string]string{"Cache-Control": "no-cache"}).Twice()
		assert.HTTPSuccess(t, c.ServeHTTP, "GET", "/api/no-cache", nil)
		assert.HTTPSuccess(t, c.ServeHTTP, "GET", "/api/no-cache", nil)
	})

	origin.AssertExpectations(t)
}

func TestCache_ClientControl(t *testing.T) {
	t.Parallel()

	origin := NewMockHandler()
	origin.StatusRoute("/api/private", http.StatusOK, nil).Twice()
	origin.StatusRoute("/api/ok", http.StatusOK, nil).Once()
	origin.StatusRoute("/api/max-stale", http.StatusOK, map[string]string{"Cache-Control": "max-age=1"})
	origin.StatusRoute("/api/min-fresh", http.StatusOK, map[string]string{"Cache-Control": "max-age=3"})

	c := cache.NewHandler(cache.NewInMemoryCache(), origin)

	t.Run("Private option enabled", func(t *testing.T) {
		// First call that uses origin server.
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/private", nil)
		req.Header.Add("Cache-Control", "private")
		c.ServeHTTP(recorder, req)

		if assert.Equal(t, http.StatusOK, recorder.Code) {
			// this one should call the origin server because the previous call shouldn't
			// be stored in the cache.
			assert.HTTPSuccess(t, c.ServeHTTP, "GET", "/api/private", nil)
		}
	})

	t.Run("Only-if-cached option enabled", func(t *testing.T) {
		t.Run("Not available response", func(t *testing.T) {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/api/not-cached", nil)
			req.Header.Add("Cache-Control", "only-if-cached")
			c.ServeHTTP(recorder, req)
			assert.Equal(t, http.StatusGatewayTimeout, recorder.Code)
		})

		t.Run("Available response", func(t *testing.T) {
			// This call should store the response in cache.
			if assert.HTTPSuccess(t, c.ServeHTTP, "GET", "/api/ok", nil) {
				// And this one should use the cache.
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest("GET", "/api/ok", nil)
				req.Header.Add("Cache-Control", "only-if-cached")
				c.ServeHTTP(recorder, req)
				assert.Equal(t, http.StatusOK, recorder.Code)
			}
		})
	})

	t.Run("Max stale option enabled", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/api/max-stale", nil)
		req.Header.Add("Cache-Control", "max-stale=2")

		// First try should fetch origin server and cache resource.
		recorder := httptest.NewRecorder()
		c.ServeHTTP(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)
		maxAgeH, ageH := recorder.Header().Get("Cache-Control"), recorder.Header().Get("Age")
		assert.Equal(t, "max-age=1", maxAgeH)
		assert.Equal(t, "0", ageH)

		// Second try should use cache resource even if the resource is stale.
		time.Sleep(2 * time.Second)
		recorder = httptest.NewRecorder()
		c.ServeHTTP(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)
		maxAgeH, ageH = recorder.Header().Get("Cache-Control"), recorder.Header().Get("Age")
		assert.Equal(t, "max-age=1", maxAgeH)
		assert.Equal(t, "2", ageH)

		// Third try should use origin server.
		time.Sleep(1 * time.Second)
		recorder = httptest.NewRecorder()
		c.ServeHTTP(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)
		maxAgeH, ageH = recorder.Header().Get("Cache-Control"), recorder.Header().Get("Age")
		assert.Equal(t, "max-age=1", maxAgeH)
		assert.Equal(t, "0", ageH)
	})

	t.Run("Min fresh option enabled", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/api/min-fresh", nil)
		req.Header.Add("Cache-Control", "min-fresh=1")

		// First try should fetch origin server and cache resource.
		recorder := httptest.NewRecorder()
		c.ServeHTTP(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)
		maxAgeH, ageH := recorder.Header().Get("Cache-Control"), recorder.Header().Get("Age")
		assert.Equal(t, "max-age=3", maxAgeH)
		assert.Equal(t, "0", ageH)

		// Second try should use cache.
		time.Sleep(1 * time.Second)
		recorder = httptest.NewRecorder()
		c.ServeHTTP(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)
		maxAgeH, ageH = recorder.Header().Get("Cache-Control"), recorder.Header().Get("Age")
		assert.Equal(t, "max-age=3", maxAgeH)
		assert.Equal(t, "1", ageH)

		// Third try should use origin server because the resource will not be available for still 2 second.
		time.Sleep(1 * time.Second)
		recorder = httptest.NewRecorder()
		c.ServeHTTP(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)
		maxAgeH, ageH = recorder.Header().Get("Cache-Control"), recorder.Header().Get("Age")
		assert.Equal(t, "max-age=3", maxAgeH)
		assert.Equal(t, "0", ageH)
	})

	origin.AssertExpectations(t)
}