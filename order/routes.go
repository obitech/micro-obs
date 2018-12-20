package order

import (
	"github.com/obitech/micro-obs/util"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var rm = util.NewRequestMetricHistogram(
	[]float64{.01, .05, .1, .25, .5, 1, 5, 10},
	[]float64{1, 5, 10, 50, 100},
)

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
			Name:        "getAllOrders",
			Method:      "GET",
			Pattern:     "/orders",
			HandlerFunc: s.getAllOrders(),
		},
		util.Route{
			Name:        "setOrder",
			Method:      "POST",
			Pattern:     "/orders",
			HandlerFunc: s.setOrder(false),
		},
		util.Route{
			Name:        "setOrder",
			Method:      "PUT",
			Pattern:     "/orders",
			HandlerFunc: s.setOrder(true),
		},
		util.Route{
			Name:        "getOrder",
			Method:      "GET",
			Pattern:     "/orders/{id:-?[0-9]+}",
			HandlerFunc: s.getOrder(),
		},
		util.Route{
			Name:        "createOrder",
			Method:      "POST",
			Pattern:     "/orders/create",
			HandlerFunc: s.createOrder(),
		},
	}

	for _, route := range routes {
		h := route.HandlerFunc

		// Logging each request
		h = util.LoggerMiddleware(h, s.logger)

		// Tracing each request
		h = util.TracerMiddleware(h, route)

		// Monitoring each request
		promHandler := util.PrometheusMiddleware(h, route, rm)

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
	promHandler := promhttp.Handler()
	// promHandler = util.TracerMiddleware(promHandler, route)
	s.router.
		Methods(route.Method).
		Path(route.Pattern).
		Name(route.Name).
		Handler(promHandler)
}
