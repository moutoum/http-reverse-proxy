package integration

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

type MockHandler struct {
	mock.Mock

	handler *http.ServeMux
}

func NewMockHandler() *MockHandler {
	return &MockHandler{
		handler: http.NewServeMux(),
	}
}

func (m *MockHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	m.handler.ServeHTTP(writer, request)
}

func (m *MockHandler) StatusRoute(pattern string, statusCode int, headers map[string]string) *mock.Call {
	var opts []func(*handler)
	for k, v := range headers {
		opts = append(opts, withHeader(k, v))
	}

	h := newHandler(&m.Mock, statusCode, opts...)
	m.handler.Handle(pattern, h)
	return h.m.On("ServeHTTP", mock.Anything, mock.Anything).Return()
}

func (m *MockHandler) DataRoute(pattern string, statusCode int, body []byte, headers map[string]string) *mock.Call {
	opts := []func(*handler){withBody(body)}
	for k, v := range headers {
		opts = append(opts, withHeader(k, v))
	}

	h := newHandler(&m.Mock, statusCode, opts...)
	m.handler.Handle(pattern, h)
	return h.m.On("ServeHTTP", mock.Anything, mock.Anything).Return()
}

type handler struct {
	m *mock.Mock

	statusCode int
	headers    map[string]string
	body       []byte
}

func withHeader(key, value string) func(h *handler) {
	return func(h *handler) {
		h.headers[key] = value
	}
}

func withBody(content []byte) func(h *handler) {
	return func(h *handler) {
		h.body = content
	}
}

func newHandler(parent *mock.Mock, statusCode int, opts ...func(*handler)) *handler {
	h := &handler{
		m:          parent,
		statusCode: statusCode,
		headers:    make(map[string]string),
		body:       nil,
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}

func (h *handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	h.m.Called(writer, request)

	writer.WriteHeader(h.statusCode)

	for k, v := range h.headers {
		writer.Header().Add(k, v)
	}

	if len(h.body) > 0 {
		_, _ = writer.Write(h.body)
	}
}
