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

func (s *Server) getAllOrders() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		span, ctx := ot.StartSpanFromContext(r.Context(), "pong")
		defer span.Finish()

		defaultErrMsg := "unable to retrieve orders"
		keys, err := s.RedisScanOrders(ctx)
		if err != nil {
			s.logger.Errorw("unable to SCAN redis for keys",
				"error", err,
			)
			s.Respond(ctx, http.StatusInternalServerError, defaultErrMsg, 0, nil, w)
			return
		}

		var orders = []*Order{}
		for _, k := range keys {
			o, err := s.RedisGetOrder(ctx, k)
			if err != nil {
				s.logger.Errorw("unable to get get order",
					"key", k,
					"error", err,
				)
				s.Respond(ctx, http.StatusInternalServerError, defaultErrMsg, 0, nil, w)
				return
			}
			orders = append(orders, o)
		}

		l := len(orders)
		if l == 0 {
			s.Respond(ctx, http.StatusNotFound, "no orders present", 0, nil, w)
			return
		}

		s.Respond(ctx, http.StatusOK, "orders retrieved", l, orders, w)
	}
}
