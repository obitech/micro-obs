package item

import (
	"net/http"

	"github.com/obitech/micro-obs/pkg/util"
)

// Routes defines a slice of all available API Routes
type Routes []util.Route

// Routes defines all HTTP routes, hanging off the main Server struct.
// Like that, all routes have access to the Server's dependencies.
func (s *Server) Routes() {
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
		h = util.LoggerWrapper(h, s.Logger)
		s.Router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(h)
	}
}
