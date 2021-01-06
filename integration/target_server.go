package integration

import (
	"net/http/httptest"
	"net/url"
)

type TargetServer struct {
	server  *httptest.Server
	handler *CustomHandler
}

func NewTargetServer() *TargetServer {
	h := NewCustomHandler()

	return &TargetServer{
		server: httptest.NewUnstartedServer(h),
		handler: h,
	}
}

func (t *TargetServer) WithRouteStatus(pattern string, statusCode int) *TargetServer {
	t.handler.StatusRoute(pattern, statusCode)
	return t
}

func (t *TargetServer) WithRouteContent(pattern string, statusCode int, content []byte) *TargetServer {
	t.handler.DataRoute(pattern, statusCode, content)
	return t
}

func (t *TargetServer) Start() *TargetServer {
	t.server.Start()
	return t
}

func (t *TargetServer) Close() {
	t.server.Close()
}

func (t *TargetServer) URL() *url.URL {
	u, _ := url.Parse(t.server.URL)
	return u
}