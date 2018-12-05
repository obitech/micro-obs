package item

import (
	"net/http"

	"github.com/obitech/micro-obs/util"
)

// Routes defines a slice of all available API Routes
type Routes []util.Route

// Routes defines all HTTP routes, hanging off the main Server struct.
// Like that, all routes have access to the Server's dependencies.
func (s *Server) createRoutes() {
	var routes = Routes{
		util.Route{
			"Root",
			"GET",
			"/",
			util.Healthz(),
		},
		util.Route{
			"Root",
			"GET",
			"/healthz",
			util.Healthz(),
		},
	}

	for _, route := range routes {
		var h http.HandlerFunc
		h = route.HandlerFunc
		h = util.LoggerWrapper(h, s.logger)
		s.router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(h)
	}
}