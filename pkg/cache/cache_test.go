package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInMemoryCache_Get(t *testing.T) {
	resource := &Resource{Status: 666}
	cache := &InMemoryCache{store: map[string]*Resource{
		"test-id": resource,
	}}

	t.Run("Fetch not present resource in store", func(t *testing.T) {
		assert.Nil(t, cache.Get("invalid-key"))
	})

	t.Run("Fetch valid resource in store", func(t *testing.T) {
		fetched := cache.Get("test-id")
		assert.Equal(t, fetched, resource)
	})

}

func TestInMemoryCache_Store(t *testing.T) {
	resource := &Resource{Status: 666}
	cache := NewInMemoryCache()
	cache.Store("test-key", resource)
	r, ok := cache.store["test-key"]
	assert.True(t, ok)
	assert.Equal(t, resource, r)
}