package proxy

import (
	"crypto/tls"
	"net/http"
)

type Option func(*Handler)

func WithInsecure() Option {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}

	return func(handler *Handler) {
		handler.transport = t
	}
}
