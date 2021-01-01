package proxy

import "net/http"

type Option func(*Handler)

func WithProxy(location, proxy string) Option {
	return func(handler *Handler) {
		handler.proxies[location] = proxy
	}
}

func WithRoundTripper(rt http.RoundTripper) Option {
	return func(handler *Handler) {
		handler.transport = rt
	}
}
