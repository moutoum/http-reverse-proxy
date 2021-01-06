package integration

import "net/http"

type CustomHandler struct {
	handler *http.ServeMux
}

func NewCustomHandler() *CustomHandler {
	return &CustomHandler{handler: http.NewServeMux()}
}

func (c *CustomHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	c.handler.ServeHTTP(writer, request)
}

func (c *CustomHandler) StatusRoute(pattern string, statusCode int) *CustomHandler {
	c.handler.HandleFunc(pattern, func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(statusCode)
	})

	return c
}

func (c *CustomHandler) DataRoute(pattern string, statusCode int, body []byte) *CustomHandler {
	c.handler.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(statusCode)
		_, _ = writer.Write(body)
	})

	return c
}
