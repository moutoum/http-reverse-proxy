package cache

import "net/http"

type Cache interface {
	Get(key string) *Resource
	Store(key string, resource *Resource)
}

type Resource struct {
	Status  int
	Headers http.Header
	Body    []byte
}

type InMemoryCache struct {
	zone map[string]*Resource
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		zone: make(map[string]*Resource, 1024),
	}
}

func (i *InMemoryCache) Get(key string) *Resource {
	resource, ok := i.zone[key]
	if !ok {
		return nil
	}

	return resource
}

func (i *InMemoryCache) Store(key string, resource *Resource) {
	i.zone[key] = resource
}