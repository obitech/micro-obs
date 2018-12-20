package order

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	ot "github.com/opentracing/opentracing-go"
)

// NotFound is a 404 Message according to the Response type
func (s *Server) notFound(w http.ResponseWriter, r *http.Request) {
	span, ctx := ot.StartSpanFromContext(r.Context(), "notFound")
	defer span.Finish()
	s.Respond(ctx, http.StatusNotFound, "resource not found", 0, nil, w)
}

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
		span, ctx := ot.StartSpanFromContext(r.Context(), "getAllOrders")
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

// setOrder creates a new Order in Redis, regardless if the items are present in the Item service.
func (s *Server) setOrder(update bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		span, ctx := ot.StartSpanFromContext(r.Context(), "setNewOrder")
		defer span.Finish()

		var (
			defaultErrMsg string
			defaultStatus int
			order         = &Order{}
		)

		switch update {
		case true:
			defaultErrMsg = "unable to update order"
			defaultStatus = http.StatusOK
		case false:
			defaultErrMsg = "unable to create order"
			defaultStatus = http.StatusCreated
		}

		// Accept payload
		body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
		if err != nil {
			s.logger.Errorw("unable to read request body",
				"error", err,
			)
			r.Body.Close()
			s.Respond(ctx, http.StatusInternalServerError, "unable to read payload", 0, nil, w)
			return
		}
		defer r.Body.Close()

		// Parse payload
		// TODO: handle multiple orders
		if err := json.Unmarshal(body, order); err != nil {
			s.logger.Errorw("unable to parse payload",
				"error", err,
			)
			s.Respond(ctx, http.StatusBadRequest, "unable to parse payload", 0, nil, w)
			return
		}

		// Check if order contains items
		if len(order.Items) == 0 {
			s.Respond(ctx, http.StatusUnprocessableEntity, "order needs items", 0, nil, w)
			return
		}

		// Check for existence
		i, err := s.RedisGetOrder(ctx, order.ID)
		if err != nil {
			s.logger.Errorw("unable to retrieve Item from Redis",
				"key", order.ID,
				"error", err,
			)
			s.Respond(ctx, http.StatusInternalServerError, defaultErrMsg, 0, nil, w)
			return
		}
		if i != nil {
			if !update {
				s.Respond(ctx, http.StatusUnprocessableEntity, fmt.Sprintf("order with ID %d already exists", order.ID), 0, nil, w)
				return
			}
		}

		// Create Order in Redis
		err = s.RedisSetOrder(ctx, order)
		if err != nil {
			s.logger.Errorw("unable to create order in redis",
				"key", order.ID,
				"error", err,
			)
			s.Respond(ctx, http.StatusInternalServerError, defaultErrMsg, 0, nil, w)
			return
		}

		s.Respond(ctx, defaultStatus, fmt.Sprintf("order %d created", order.ID), 1, []*Order{order}, w)
	}
}

func (s *Server) createOrder() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		span, ctx := ot.StartSpanFromContext(r.Context(), "createOrder")
		defer span.Finish()

		// Accept payload
		body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
		if err != nil {
			s.logger.Errorw("unable to read request body",
				"error", err,
			)
			r.Body.Close()
			s.Respond(ctx, http.StatusInternalServerError, "unable to read payload", 0, nil, w)
			return
		}
		defer r.Body.Close()

		// Parse payload
		var order *Order
		if err := json.Unmarshal(body, &order); err != nil {
			s.logger.Errorw("unable to parse payload",
				"error", err,
			)
			s.Respond(ctx, http.StatusBadRequest, "unable to parse payload", 0, nil, w)
			return
		}

		// Check if order contains items
		if len(order.Items) == 0 {
			s.Respond(ctx, http.StatusUnprocessableEntity, "order needs items", 0, nil, w)
			return
		}

		// Get requested items from item service
		// TODO: Send & process in bulk
		for _, orderItem := range order.Items {
			itemItem, err := s.getItem(ctx, orderItem.ID)
			if err != nil {
				if err, ok := err.(notFoundError); ok {
					s.Respond(ctx, http.StatusNotFound, err.Error(), 0, nil, w)
					return
				}

				s.logger.Errorw("unable to retrieve item from item service",
					"itemID", orderItem.ID,
					"error", err,
				)
				s.Respond(ctx, http.StatusInternalServerError, "unable to retrieve item from item service", 0, nil, w)
				return
			}

			if itemItem.Qty < orderItem.Qty {
				msg := fmt.Sprintf("not enough units of %s available (%d avail, %d requested)", orderItem.ID, orderItem.Qty, itemItem.Qty)
				s.Respond(ctx, http.StatusUnprocessableEntity, msg, 0, nil, w)
				return
			}
		}

		// Get OrderID from Redis
		id, err := s.RedisGetNextOrderID(ctx)
		if err != nil {
			s.logger.Errorw("unable to get next order ID",
				"error", err,
			)
			s.Respond(ctx, http.StatusInternalServerError, "unable to create order", 0, nil, w)
			return
		}
		order.ID = id

		// Create order
		err = s.RedisSetOrder(ctx, order)
		if err != nil {
			s.logger.Errorw("unable to create order in redis",
				"error", err,
			)
		}

		// Respond
		msg := fmt.Sprintf("order %d created", order.ID)
		s.Respond(ctx, http.StatusCreated, msg, 1, []*Order{order}, w)
	}
}

func (s *Server) getOrder() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		span, ctx := ot.StartSpanFromContext(r.Context(), "getOrder")
		defer span.Finish()

		pr := mux.Vars(r)
		id, err := strconv.ParseInt(pr["id"], 10, 64)
		if err != nil {
			s.Respond(ctx, http.StatusBadRequest, "unable to parse ID", 0, nil, w)
			return
		}

		order, err := s.RedisGetOrder(ctx, id)
		if err != nil {
			s.logger.Errorw("unable to get key from redis",
				"key", id,
				"error", err,
			)
			s.Respond(ctx, http.StatusInternalServerError, "unable to retreive item", 0, nil, w)
			return
		}

		if order == nil {
			s.Respond(ctx, http.StatusNotFound, fmt.Sprintf("order %d doesn't exist", id), 0, nil, w)
			return
		}

		s.Respond(ctx, http.StatusOK, "order retrieved", 1, []*Order{order}, w)
	}
}
