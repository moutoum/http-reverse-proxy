package server

import (
	"context"
	"net/http"

	"github.com/moutoum/http-reverse-proxy/pkg/proxy"
)

type Server struct {
	s *http.Server
}

func New(opts ...Option) *Server {
	s := &Server{
		s: &http.Server{
			Handler: &proxy.Proxy{},
		},
	}

	for _, o := range opts {
		o(s)
	}

	return s
}

func (s *Server) Serve() error {
	return s.s.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.s.Shutdown(ctx)
}

func (s *Server) Close() error {
	return s.s.Close()
}
