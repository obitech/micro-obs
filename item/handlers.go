package item

import (
	"net/http"

	"github.com/obitech/micro-obs/util"
)

// pong sends a simple JSON response.
func (s *Server) pong() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if res, err := util.NewResponse(200, "pong", 0, nil); err != nil {
			s.logger.Errorf("sending JSON response failed",
				"error", err,
				"response", res,
			)
		}
	}
}

// getAllItems retrieves all items from Redis.
func (s *Server) getAllKeys() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
