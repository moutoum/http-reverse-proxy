package integration

import (
	"net/http"
	"net/http/httptest"
	"net/url"
)

type OriginServer struct {
	server  *httptest.Server
	handler *http.ServeMux
}

func NewOriginServer() *OriginServer {
	h := http.NewServeMux()

	return &OriginServer{
		server: httptest.NewUnstartedServer(h),
		handler: h,
	}
}

func (t *OriginServer) WithRouteStatus(pattern string, statusCode int) *OriginServer {
	t.handler.HandleFunc(pattern, func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(statusCode)
	})

	return t
}

func (t *OriginServer) WithRouteContent(pattern string, statusCode int, content []byte) *OriginServer {
	t.handler.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(statusCode)
		_, _ = writer.Write(content)
	})

	return t
}

func (t *OriginServer) Start() *OriginServer {
	t.server.Start()
	return t
}

func (t *OriginServer) Close() {
	t.server.Close()
}

func (t *OriginServer) URL() *url.URL {
	u, _ := url.Parse(t.server.URL)
	return u
}