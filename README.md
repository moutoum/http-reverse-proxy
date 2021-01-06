# HTTP Reverse Proxy

This project is a small http reverse proxy written in Golang.
The goal is to learn and implement a http server proxy that embeds
a cache feature.

This server is very primitive and does not implements the full-featured
behavior for proxies and caching mechanisms.

## Proxy server

This repo provides a cli binary that runs the proxy server with the
cache. Several options are available.

#### Build

```shell script
go build ./cmd/proxy-server/proxy-server.go
```

#### Help

```shell script
./proxy-server --help
```

#### Run

```shell script
./proxy-server --target-server "http://localhost:5051" --bind-addr ":5050"
```

## Features

- Can proxy not secure http requests to a http server.
- Cache all GET and HEAD requests.
- Several `Cache-Control" options (max-age, min-fresh, public, max-stale, ...)

## Resources

- https://www.digitalocean.com/community/tutorials/web-caching-basics-terminology-http-headers-and-caching-strategies
- https://github.com/lox/httpcache
- https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control
- https://developer.mozilla.org/en-US/docs/Web/HTTP/Caching
- https://developer.mozilla.org/en-US/docs/Web/HTTP/Status
- https://github.com/traefik/traefik