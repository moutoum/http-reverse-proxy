package main

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"os/signal"

	"github.com/moutoum/http-reverse-proxy/pkg/proxy"
	"github.com/sirupsen/logrus"
)

func main() {
	p := proxy.New(&url.URL{
		Scheme: "http",
		Host:   ":5051",
		Path:   "/api",
	})

	s := http.Server{
		Addr: ":5050",
		Handler: p,
	}

	go func() {
		logrus.Info("Start listening on port 5050")
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.WithError(err).Error("Error while serving HTTP server")
			return
		}

	}()

	c := make(chan os.Signal)
	closingChan := make(chan interface{}, 1)

	signal.Notify(c, os.Interrupt)

	// Wait for interrupt.
	<-c

	ctx, cancel := context.WithCancel(context.Background())

	// Try to gracefully shut down the pending requests.
	go func() {
		logrus.Info("Waiting for connection to exit")
		if err := s.Shutdown(ctx); err != nil {
			logrus.WithError(err).Error("Error while shutting down server")
			return
		}

		close(closingChan)
	}()

	select {
	case <-closingChan:
		// Gracefully attempt worked.
		cancel()
		logrus.Info("Shutting down. Bye")

	case <-c:
		// We couldn't wait more time, and a second interrupt occurred
		// to force stop the server.
		cancel()
		_ = s.Close()
		logrus.Info("Force shutdown.")
	}
}
