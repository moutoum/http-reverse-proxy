package proxy

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
)

type Handler struct {
	target    *url.URL
	transport http.RoundTripper
}

var _ http.Handler = (*Handler)(nil)

func New(target *url.URL, opts ...Option) *Handler {
	h := &Handler{
		transport: http.DefaultTransport,
		target:    target,
	}

	for _, o := range opts {
		o(h)
	}

	return h
}

func (h *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	outgoingRequest := request.Clone(request.Context())
	outgoingRequest.URL = mergeURLs(request.URL, h.target)

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
	headers := writer.Header()
	for key, values := range response.Header {
		for _, value := range values {
			headers.Add(key, value)
		}
	}

	// Forward response body.
	if _, err := io.Copy(writer, response.Body); err != nil {
		return err
	}

	return nil
}

func mergeURLs(req, target *url.URL) *url.URL {
	u := *target

	if !strings.HasSuffix(u.Path, "/") {
		u.Path += "/"
	}
	u.Path += strings.TrimPrefix(req.Path, "/")

	if len(u.RawQuery) == 0 || len(req.RawQuery) == 0 {
		u.RawQuery = u.RawQuery + req.RawQuery
	} else {
		u.RawQuery = u.RawQuery + "&" + req.RawQuery
	}

	return &u
}
