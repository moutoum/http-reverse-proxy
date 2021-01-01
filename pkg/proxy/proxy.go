package proxy

import (
	"io"
	"net/http"

	"github.com/sirupsen/logrus"
)

type Handler struct {
	proxies   map[string]string
	transport http.RoundTripper
}

var _ http.Handler = (*Handler)(nil)

func New(opts ...Option) *Handler {
	h := &Handler{
		proxies: make(map[string]string),
		transport: http.DefaultTransport,
	}

	for _, o := range opts {
		o(h)
	}

	return h
}

func (h *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	outgoingRequest := request.Clone(request.Context())

	host, ok := h.proxies[request.RequestURI]
	if !ok {
		writer.WriteHeader(http.StatusBadGateway)
		return
	}

	outgoingRequest.URL.Scheme = "http"
	outgoingRequest.URL.Host = host

	response, err := h.transport.RoundTrip(outgoingRequest)
	if err != nil {
		logrus.WithError(err).Error("Error while sending request")
		writer.WriteHeader(http.StatusBadGateway)
		return
	}

	if err = copyResponse(response, writer); err != nil {
		logrus.WithError(err).Error("Error while copying response")
		writer.WriteHeader(http.StatusBadGateway)
		return
	}
}

func copyResponse(response *http.Response, writer http.ResponseWriter) error {
	// Forward response status code.
	writer.WriteHeader(response.StatusCode)

	// Forward http headers.
	for key, values := range response.Header {
		for _, value := range values {
			writer.Header().Add(key, value)
		}
	}

	// Forward response body.
	if _, err := io.Copy(writer, response.Body); err != nil {
		return err
	}

	return nil
}
