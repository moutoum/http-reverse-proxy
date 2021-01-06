package main

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"os/signal"

	"github.com/moutoum/http-reverse-proxy/pkg/cache"
	"github.com/moutoum/http-reverse-proxy/pkg/proxy"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// URLGenericValue helps to parse the CLI
// argument into an url.URL.
//
// It implements the cli.Generic interface.
type URLGenericValue struct {
	url *url.URL
}

// Set is the "cli.Generic" interface implementation.
func (u *URLGenericValue) Set(value string) error {
	v, err := url.Parse(value)
	if err != nil {
		return err
	}

	u.url = v
	return nil
}

// String is the "cli.Generic" interface implementation.
func (u *URLGenericValue) String() string {
	if u.url == nil {
		return ""
	}
	return u.url.String()
}

func main() {
	app := &cli.App{
		Name:        "proxy-server",
		Usage:       "HTTP proxy server",
		HideVersion: true,
		Authors: []*cli.Author{{
			Name:  "Maxence Moutoussamy",
			Email: "maxence.moutoussamy1@gmail.com",
		}},
		Action: app,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "bind-addr",
				Aliases: []string{"b"},
				Usage:   "Binding address for the proxy server",
				Value:   ":80",
			},
			&cli.GenericFlag{
				Name:     "target-server",
				Aliases:  []string{"t"},
				Usage:    "Target server URL to use to forward requests",
				Required: true,
				Value:    &URLGenericValue{},
			},
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"d"},
				Usage:   "Enable debug mode",
				Value:   false,
			},
			&cli.BoolFlag{
				Name:    "enable-cache",
				Aliases: []string{"c"},
				Usage:   "Enable the cache feature",
				Value:   false,
			},
		},
	}

	_ = app.Run(os.Args)
}

func app(args *cli.Context) error {
	if args.Bool("debug") {
		logrus.SetLevel(logrus.DebugLevel)
	}

	urlValue := args.Generic("target-server").(*URLGenericValue)
	var h http.Handler = proxy.New(urlValue.url)
	if args.Bool("enable-cache") {
		h = cache.NewHandler(cache.NewInMemoryCache(), h)
	}
	s := http.Server{
		Addr:    args.String("bind-addr"),
		Handler: h,
	}

	go func() {
		logrus.Infof("Start listening at %s", args.String("bind-addr"))
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

	return nil
}
