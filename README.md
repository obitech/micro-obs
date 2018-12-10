# [micro-obs](https://github.comobitech/microservices-observability)

[![Build Status](https://travis-ci.org/obitech/micro-obs.svg?branch=master)](https://travis-ci.org/obitech/micro-obs) [![Go Report Card](https://goreportcard.com/badge/github.com/obitech/micro-obs)](https://goreportcard.com/report/github.com/obitech/micro-obs)

Demonstrating monitoring, logging and tracing of a simple microservices shop application on Kubernetes.

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

Run:

```bash
docker-compose up -d
```

### Binary

Start redis:

```bash
docker-compose up -d redis
```

Start the item server locally:

```bash
./bin/item
{"level":"info","ts":1544472652.93183,"msg":"Testing redis connection"}
{"level":"info","ts":1544472652.9464881,"msg":"Server listening","address":":8080","endpoint":"127.0.0.1:8080"}
```

## Use it

Do some `curl`s:

```bash
for i in {1..5}; do curl http://localhost:8080/healthz; done 
```

Check the metrics:

```bash
curl localhost:8080/metrics
# ...
# HELP in_flight_requests A gauge of requests currently being served by the wrapped handler.
# TYPE in_flight_requests gauge
in_flight_requests 0
# HELP promhttp_metric_handler_requests_in_flight Current number of scrapes being served.
# TYPE promhttp_metric_handler_requests_in_flight gauge
promhttp_metric_handler_requests_in_flight 1
# HELP promhttp_metric_handler_requests_total Total number of scrapes by HTTP status code.
# TYPE promhttp_metric_handler_requests_total counter
promhttp_metric_handler_requests_total{code="200"} 0
promhttp_metric_handler_requests_total{code="500"} 0
promhttp_metric_handler_requests_total{code="503"} 0
# HELP request_duration_seconds A histogram for latencies for requests.
# TYPE request_duration_seconds histogram
request_duration_seconds_bucket{handler="healthz",method="get",le="0.01"} 5
request_duration_seconds_bucket{handler="healthz",method="get",le="0.05"} 5
request_duration_seconds_bucket{handler="healthz",method="get",le="0.1"} 5
request_duration_seconds_bucket{handler="healthz",method="get",le="0.25"} 5
request_duration_seconds_bucket{handler="healthz",method="get",le="0.5"} 5
request_duration_seconds_bucket{handler="healthz",method="get",le="1"} 5
request_duration_seconds_bucket{handler="healthz",method="get",le="5"} 5
request_duration_seconds_bucket{handler="healthz",method="get",le="10"} 5
request_duration_seconds_bucket{handler="healthz",method="get",le="+Inf"} 5
request_duration_seconds_sum{handler="healthz",method="get"} 0.005152529
request_duration_seconds_count{handler="healthz",method="get"} 5
# HELP response_size_bytes A histogram of response sizes for requests.
# TYPE response_size_bytes histogram
response_size_bytes_bucket{le="1"} 0
response_size_bytes_bucket{le="5"} 5
response_size_bytes_bucket{le="10"} 5
response_size_bytes_bucket{le="50"} 5
response_size_bytes_bucket{le="100"} 5
response_size_bytes_bucket{le="+Inf"} 5
response_size_bytes_sum 15
response_size_bytes_count 5
```

## [item](https://godoc.org/github.com/obitech/micro-obs/item)
[![godoc reference for item](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/obitech/micro-obs/item) 

## [util](https://godoc.org/github.com/obitech/micro-obs/item)
[![godoc reference for item](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/obitech/micro-obs/util) 

## License

[MIT](https://choosealicense.com/licenses/mit/#)