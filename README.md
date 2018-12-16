# [micro-obs](https://github.comobitech/microservices-observability)

[![Build Status](https://travis-ci.org/obitech/micro-obs.svg?branch=master)](https://travis-ci.org/obitech/micro-obs) [![Go Report Card](https://goreportcard.com/badge/github.com/obitech/micro-obs)](https://goreportcard.com/report/github.com/obitech/micro-obs) [![](https://img.shields.io/docker/automated/jrottenberg/ffmpeg.svg)](https://hub.docker.com/r/obitech/micro-obs)


Demonstrating monitoring, logging and tracing of a simple microservices shop application on Kubernetes:

- Structured logging via [zap](https://github.com/uber-go/zap)
- Automatic endpoint monitoring exposing metrics to [Prometheus](https://github.com/prometheus/prometheus)
- Internal & distributed tracing via [Jaeger](https://github.com/jaegertracing/jaeger)

## Build it

To build it from source you need [Go 1.11+](https://golang.org/dl/) installed.

This project uses [Go Modules](https://github.com/golang/go/wiki/Modules) so you can clone the repo to anywhere:

```bash
git clone https://github.com/obitech/micro-obs.git
cd micro-obs/
```

Run `make` to test & build it:

```bash
make
# ...
```

Or `make docker` to build a Docker image:

```bash
make docker TAG=yourname/yourimage
# ...
```

## Run it

### Docker
```bash
cd deploy/docker
docker-compose up -d
```

Service|Location
---|---
Item API|http://localhost:8080/
Jaeger UI|http://localhost:16686/
Prometheus|http://localhost:9090/
Grafana|http://localhost:3000/

## [item](https://godoc.org/github.com/obitech/micro-obs/item)
[![godoc reference for item](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/obitech/micro-obs/item) 


## [util](https://godoc.org/github.com/obitech/micro-obs/item)
[![godoc reference for item](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/obitech/micro-obs/util) 

## License

[MIT](https://choosealicense.com/licenses/mit/#)