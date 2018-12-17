package order

import (
	_ "encoding/json"
	_ "fmt"
	_ "io"
	_ "io/ioutil"
	"net/http"

	_ "github.com/gorilla/mux"
	ot "github.com/opentracing/opentracing-go"
)

// pong sends a simple JSON response.
func (s *Server) pong() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		span, ctx := ot.StartSpanFromContext(r.Context(), "pong")
		defer span.Finish()
		s.Respond(ctx, http.StatusOK, "pong", 0, nil, w)
	}
}
