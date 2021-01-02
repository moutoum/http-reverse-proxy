package proxy

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
)

// Handler represents a proxy server.
type Handler struct {
	target    *url.URL
	transport http.RoundTripper
}

// Static implementation checker.
var _ http.Handler = (*Handler)(nil)

// New creates a proxy that can be served as a http.Handler.
// It takes the target server to forward requests.
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

// ServeHTTP exposes the configured proxy.
//
// The function forwards the incoming request to the target server provided
// with the type and then reads the HTTP response.
// The response is forwarded to the client connection.
// If an error occurs during the forwarding process, it sends back a
// 502 Bad Gateway status to the client.
//
// ServeHTTP is the `http.Handler` implementation for the `Handler` type.
func (h *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	outgoingRequest := request.Clone(request.Context())
	outgoingRequest.URL = mergeURLs(request.URL, h.target)

	// Sends the request to the target server.
	// Note: Cannot use a simple `http.Client` because the implementation
	//       returns error with HTTP semantic errors (4xx, 5xx, ...).
	response, err := h.transport.RoundTrip(outgoingRequest)
	if err != nil {
		logrus.WithError(err).Error("Error while sending request")
		writer.WriteHeader(http.StatusBadGateway)
		return
	}

	defer response.Body.Close()

	if err = copyResponse(response, writer); err != nil {
		logrus.WithError(err).Error("Error while copying response")
		writer.WriteHeader(http.StatusBadGateway)
		return
	}
}

// copyResponse forwards the given response to the response writer.
//
// It copies the HTTP status code, merges the headers and copies the
// response content.
// The error could be non-nil if it wasn't able to copy the body to
// the response writer.
//
// NOTE: For the headers, the values are mixed together, it means that
//       if the writer already has some headers, they will not be erased.
//       Instead, the header will contains all the values for this key.
func copyResponse(response *http.Response, writer http.ResponseWriter) error {
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

	// Forward response status code.
	writer.WriteHeader(response.StatusCode)

	return nil
}

// mergeURLs joins the request and the target URLs together
// to allow URI composition.
// For a request that requests "/entities" and a target that
// handles "/api", the generated URL will be "/api/entities".
//
// NOTE: It also merges the query parameters.
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
