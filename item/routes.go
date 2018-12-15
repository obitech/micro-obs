package item

import (
	"github.com/obitech/micro-obs/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var rm = util.NewRequestMetricHistogram(
	[]float64{.01, .05, .1, .25, .5, 1, 5, 10},
	[]float64{1, 5, 10, 50, 100},
)

func init() {
	prometheus.MustRegister(rm.InFlightGauge, rm.Counter, rm.Duration, rm.ResponseSize)
}

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
	}

	for _, route := range routes {
		h := route.HandlerFunc

		// Logging each request
		h = util.LoggerWrapper(h, s.logger)

		// Closure chain to wrap HandlerFunc with monitoring capabilities
		promHandler := promhttp.InstrumentHandlerInFlight(
			rm.InFlightGauge,
			promhttp.InstrumentHandlerDuration(
				rm.Duration.MustCurryWith(prometheus.Labels{"handler": route.Name}),
				promhttp.InstrumentHandlerCounter(
					rm.Counter,
					promhttp.InstrumentHandlerResponseSize(
						rm.ResponseSize,
						h,
					),
				),
			),
		)

		s.router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(promHandler)
	}

	// Prometheus endpoint
	s.router.
		Methods("GET").
		Path("/metrics").
		Name("metrics").
		Handler(util.LoggerWrapper(promhttp.Handler(), s.logger))
}
