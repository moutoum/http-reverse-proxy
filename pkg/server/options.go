package server

import "crypto/tls"

type Option func(*Server)

func WithAddr(addr string) Option {
	return func(server *Server) {
		server.s.Addr = addr
	}
}

func WithTLS(config *tls.Config) Option {
	return func(server *Server) {
		server.s.TLSConfig = config
	}
}
