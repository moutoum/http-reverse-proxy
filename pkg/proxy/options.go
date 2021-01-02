package proxy

import (
	"net/http"
)

type Option func(*Handler)

func WithRoundTripper(rt http.RoundTripper) Option {
	return func(handler *Handler) {
		handler.transport = rt
	}
}
