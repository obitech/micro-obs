package item

import (
	"github.com/obitech/micro-obs/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var rm = util.NewRequestMetricHistogram(
	append([]float64{.001, .003}, prometheus.DefBuckets...),
	[]float64{1, 5, 10, 50, 100},
)

// func init() {
// 	prometheus.MustRegister(rm.InFlightGauge, rm.Counter, rm.Duration, rm.ResponseSize)
// 	// prometheus.Register(rm.InFlightGauge)
// 	// prometheus.Register(rm.Counter)
// 	// prometheus.Register(rm.Duration)
// 	// prometheus.Register(rm.ResponseSize)
// }

// Routes defines all HTTP routes, hanging off the main Server struct.
// Like that, all routes have access to the Server's dependencies.
func (s *Server) createRoutes() {
	var routes = util.Routes{
		util.Route{
			Name:        "pong",
			Method:      "GET",
			Pattern:     "/",
			HandlerFunc: s.pong(),
		},
		util.Route{
			Name:        "healthz",
			Method:      "GET",
			Pattern:     "/healthz",
			HandlerFunc: util.Healthz(),
		},
		util.Route{
			Name:        "getAllItems",
			Method:      "GET",
			Pattern:     "/items",
			HandlerFunc: s.getAllItems(),
		},
		util.Route{
			Name:        "setItemsPOST",
			Method:      "POST",
			Pattern:     "/items",
			HandlerFunc: s.setItem(false),
		},
		util.Route{
			Name:        "setItemsPUT",
			Method:      "PUT",
			Pattern:     "/items",
			HandlerFunc: s.setItem(true),
		},
		util.Route{
			Name:        "getItem",
			Method:      "GET",
			Pattern:     "/items/{id:[a-zA-Z0-9]+}",
			HandlerFunc: s.getItem(),
		},
		util.Route{
			Name:        "delItem",
			Method:      "DELETE",
			Pattern:     "/items/{id:[a-zA-Z0-9]+}",
			HandlerFunc: s.delItem(),
		},
		util.Route{
			Name:        "delay",
			Method:      "GET",
			Pattern:     "/delay",
			HandlerFunc: s.delay(),
		},
		util.Route{
			Name:        "simulateError",
			Method:      "GET",
			Pattern:     "/error",
			HandlerFunc: s.simulateError(),
		},
	}

	for _, route := range routes {
		h := route.HandlerFunc

		// Logging each request
		h = util.LoggerMiddleware(h, s.logger)

		// Tracing each request
		h = util.TracerMiddleware(h, route)

		// Monitoring each request
		promHandler := util.PrometheusMiddleware(h, route.Name, rm)

		s.router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(promHandler)
	}

	// Prometheus endpoint
	route := util.Route{
		Name:        "metrics",
		Method:      "GET",
		Pattern:     "/metrics",
		HandlerFunc: nil,
	}

	promHandler := promhttp.HandlerFor(s.promReg, promhttp.HandlerOpts{})
	promHandler = promhttp.InstrumentMetricHandler(s.promReg, promHandler)
	s.router.
		Methods(route.Method).
		Path(route.Pattern).
		Name(route.Name).
		Handler(promHandler)

	// 404 handler
	notFound := util.PrometheusMiddleware(s.notFound(), "metrics", rm)
	s.router.NotFoundHandler = notFound
}
