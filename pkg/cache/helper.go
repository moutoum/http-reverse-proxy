package cache

import "net/http"

type ResourceWriter struct {
	writer http.ResponseWriter
	status int
	body   []byte
}

func NewResourceWriter(upstream http.ResponseWriter) *ResourceWriter {
	return &ResourceWriter{
		writer: upstream,
	}
}

func (r *ResourceWriter) Header() http.Header {
	return r.writer.Header()
}

func (r *ResourceWriter) Write(bytes []byte) (int, error) {
	_, err := r.writer.Write(bytes)
	if err != nil {
		return 0, err
	}

	r.body = bytes

	if r.status == 0 {
		r.status = http.StatusOK
	}

	return len(bytes), nil
}

func (r *ResourceWriter) WriteHeader(statusCode int) {
	r.writer.WriteHeader(statusCode)
	r.status = statusCode
}

func (r *ResourceWriter) Resource() *Resource {
	return &Resource{
		Status:  r.status,
		Headers: r.writer.Header(),
		Body:    r.body,
	}
}
