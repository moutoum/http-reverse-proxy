package cache

import (
	"net/http"
	"sync"
	"time"
)

// Cache is an interface that provides a cache storing
// behavior.
type Cache interface {

	// Get returns a stored resource.
	// The resource can be nil, and in this case, it means
	// that the resource is not found in the storage.
	Get(key string) *Resource

	// Store saves a resource with the provided key.
	Store(key string, resource *Resource)
}

// Resource represents a cache entry that stores the
// cached response status.
type Resource struct {

	// Status contains the HTTP status code.
	Status int

	// Headers contains the HTTP headers.
	Headers http.Header

	// Body is the HTTP response's content raw bytes.
	Body []byte

	// Date when the resource was created.
	Date time.Time

	// cc is the cache control parsed header.
	cc *CacheControl
}

// Age returns the current age of the resource depending on
// the current Date and the resource's creation Date.
func (r *Resource) Age() time.Duration {
	return time.Now().Sub(r.Date)
}

// InMemoryCache is a cache that stores the resources in the
// program memory.
type InMemoryCache struct {

	// store is a map that represents the resources indexed by
	// a key.
	store *sync.Map
}

// NewInMemoryCache creates an in memory cache.
func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		store: &sync.Map{},
	}
}

// Get is the `Cache` interface implementation.
func (i *InMemoryCache) Get(key string) *Resource {
	resource, ok := i.store.Load(key)
	if !ok {
		return nil
	}

	return resource.(*Resource)
}

// Store is the `Cache` interface implementation.
func (i *InMemoryCache) Store(key string, resource *Resource) {
	i.store.Store(key, resource)
}
