package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

type LogMiddleware struct {
	handler http.Handler
}

func (l *LogMiddleware) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	logrus.WithFields(logrus.Fields{
		"remote-addr": request.Header.Get("X-Proxy-Remote-Addr"),
	}).Infof("Received %s %s", request.Method, request.RequestURI)

	l.handler.ServeHTTP(writer, request)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/hello-world", func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintf(writer, "Hello from %s\n", os.Args[0])
	})

	mux.HandleFunc("/api/max-age", func(writer http.ResponseWriter, request *http.Request) {
		age := request.FormValue("t")
		writer.Header().Set("Cache-Control", "max-age="+age)
		_, _ = fmt.Fprintf(writer, "This a resource with max-age=%s\n", age)
	})

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		fmt.Println("Starting listening on port 8080")
		if err := http.ListenAndServe(":8080", &LogMiddleware{handler: mux}); err != nil {
			fmt.Printf("An error occurred while serving http server: %v\n", err)
		}
		wg.Done()
	}()

	go func() {
		fmt.Println("Starting secured listening on port 8443")
		if err := http.ListenAndServeTLS(":8443", "./certs/server.crt", "./certs/server.key", &LogMiddleware{handler: mux}); err != nil {
			fmt.Printf("An error occurred while serving http server: %v\n", err)
		}
		wg.Done()
	}()

	wg.Wait()
}
