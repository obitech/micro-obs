package util

import (
	"io"
	"net/http"

	opentracing "github.com/opentracing/opentracing-go"
	config "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics/prometheus"
)

// InitTracer returns an instance of Jaeger Tracer that samples 100% of traces and logs all spans to stdout.
func InitTracer(service string, logger *Logger) (opentracing.Tracer, io.Closer, error) {
	cfg := &config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans: true,
		},
	}
	tracer, closer, err := cfg.New(
		service,
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
		tracer := opentracing.GlobalTracer()
		span := tracer.StartSpan("request")
		ctx := opentracing.ContextWithSpan(r.Context(), span)
		defer span.Finish()

		span.SetTag("method", r.Method)
		span.SetTag("url", r.URL.Path)
		span.SetTag("handler", route.Name)

		r = r.WithContext(ctx)
		inner.ServeHTTP(w, r)
	})
}
