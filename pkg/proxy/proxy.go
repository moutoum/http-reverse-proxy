package proxy

import (
	"net/http"
)

type Proxy struct {}

func (p *Proxy) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	panic("implement me")
}



