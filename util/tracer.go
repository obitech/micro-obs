package util

import (
	"context"
	"fmt"
	"io"
	"net/http"

	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	config "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics/prometheus"
)

// InitTracer returns an instance of Jaeger Tracer that samples 100% of traces and logs all spans to stdout.
func InitTracer(serviceName string, logger *Logger) (ot.Tracer, io.Closer, error) {
	cfg, err := config.FromEnv()
	if err != nil {
		return nil, nil, err
	}

	cfg.Sampler.Type = "const"
	cfg.Sampler.Param = 1
	cfg.Reporter.LogSpans = true

	tracer, closer, err := cfg.New(
		serviceName,
		config.Logger(logger),
		config.Metrics(prometheus.New()),
	)
	if err != nil {
		return nil, nil, err
	}
	return tracer, closer, nil
}

// TracerMiddleware adds a Span to the request Context ready for other handlers to use it.
func TracerMiddleware(inner http.Handler, route Route) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ctx context.Context
		var span ot.Span
		tracer := ot.GlobalTracer()

		// If possible, extract span context from headers
		spanCtx, _ := tracer.Extract(ot.HTTPHeaders, ot.HTTPHeadersCarrier(r.Header))
		if spanCtx == nil {
			span = tracer.StartSpan("request")
			defer span.Finish()
			ctx = ot.ContextWithSpan(r.Context(), span)
		} else {
			span = tracer.StartSpan("request", ext.RPCServerOption(spanCtx))
			defer span.Finish()
			ctx = ot.ContextWithSpan(r.Context(), span)
		}
		for k, v := range r.Header {
			span.SetTag(fmt.Sprintf("header.%s", k), v)
		}
		span.SetTag("method", r.Method)
		span.SetTag("url", r.URL.Path)
		span.SetTag("handler", route.Name)

		r = r.WithContext(ctx)
		inner.ServeHTTP(w, r)
	})
}
