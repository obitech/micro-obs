package item

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
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

// getAllItems retrieves all items from Redis.
func (s *Server) getAllItems() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		span, ctx := ot.StartSpanFromContext(r.Context(), "getAllItems")
		defer span.Finish()

		defaultErrMsg := "unable to retrieve items"
		keys, err := s.RedisScanKeys(ctx)
		if err != nil {
			s.logger.Errorw("unable to SCAN redis for keys",
				"error", err,
			)
			s.Respond(ctx, http.StatusInternalServerError, defaultErrMsg, 0, nil, w)
			return
		}

		var items = []*Item{}
		for _, k := range keys {
			i, err := s.RedisGetItem(ctx, k)
			if err != nil {
				s.logger.Errorw("unable to retrieve item",
					"key", k,
					"error", err,
				)
				s.Respond(ctx, http.StatusInternalServerError, defaultErrMsg, 0, nil, w)
				return
			}
			items = append(items, i)
		}

		l := len(items)
		if l == 0 {
			s.Respond(ctx, http.StatusNotFound, "no items present", 0, nil, w)
			return
		}

		s.Respond(ctx, http.StatusOK, "items retrieved", l, items, w)
	}
}

// setItem sets an Item as a hash in Redis with URL parameters.
func (s *Server) setItem(update bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		span, ctx := ot.StartSpanFromContext(r.Context(), "setItem")
		defer span.Finish()

		var (
			defaultErrMsg string
			item          = &Item{}
			status        int
		)

		switch update {
		case true:
			defaultErrMsg = "unable to update item"
			status = http.StatusOK
		case false:
			defaultErrMsg = "unable to create item"
			status = http.StatusCreated
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
		if err := json.Unmarshal(body, item); err != nil {
			s.logger.Errorw("unable to parse payload",
				"error", err,
			)
			s.Respond(ctx, http.StatusUnprocessableEntity, "unable to parse payload", 0, nil, w)
			return
		}

		// Catch empty response
		if item.ID == "" {
			s.Respond(ctx, http.StatusUnprocessableEntity, "invalid data", 0, nil, w)
			return
		}

		if err := item.SetID(ctx); err != nil {
			s.logger.Errorw("unable to set item ID",
				"error", err,
			)
			s.Respond(ctx, http.StatusInternalServerError, defaultErrMsg, 0, nil, w)
			return
		}
		s.logger.Debugw("item struct created",
			"id", item.ID,
			"name", item.Name,
			"desc", item.Desc,
			"qty", item.Qty,
		)

		// Check for key existence
		i, err := s.RedisGetItem(ctx, item.ID)
		if err != nil {
			s.logger.Errorw("unable to retrieve Item from Redis",
				"key", item.ID,
				"error", err,
			)
			s.Respond(ctx, http.StatusInternalServerError, defaultErrMsg, 0, nil, w)
			return
		}
		if i != nil {
			if !update {
				s.Respond(ctx, http.StatusOK, fmt.Sprintf("item with name %s already exists", item.Name), 0, nil, w)
				return
			}
		}

		// Create Item in Redis
		err = s.RedisSetItem(ctx, item)
		if err != nil {
			s.logger.Errorw("unable to create Item in redis",
				"key", item.ID,
				"error", err,
			)
			s.Respond(ctx, http.StatusInternalServerError, defaultErrMsg, 0, nil, w)
			return
		}

		s.Respond(ctx, status, fmt.Sprintf("item %s created", item.Name), 1, []*Item{item}, w)
	}
}

// getItem retrieves a single Item by ID from Redis.
func (s *Server) getItem() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		span, ctx := ot.StartSpanFromContext(r.Context(), "getItem")
		defer span.Finish()

		pr := mux.Vars(r)
		key := pr["id"]

		item, err := s.RedisGetItem(ctx, key)
		if err != nil {
			s.logger.Errorw("unable to get key from redis",
				"key", key,
				"error", err,
			)
		}
		if item == nil {
			s.Respond(r.Context(), http.StatusNotFound, fmt.Sprintf("item with ID %s doesn't exist", key), 0, nil, w)
			return
		}
		s.Respond(r.Context(), http.StatusOK, "item retrieved", 1, []*Item{item}, w)
	}
}

// delItem deletes a single item by ID.
func (s *Server) delItem() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		span, ctx := ot.StartSpanFromContext(r.Context(), "delItem")
		defer span.Finish()

		pr := mux.Vars(r)
		key := pr["id"]

		err := s.RedisDelItem(ctx, key)
		if err != nil {
			s.logger.Errorw("unable to delete key from redis",
				"key", key,
				"error", err,
			)
			s.Respond(r.Context(), http.StatusInternalServerError, "an error occured while tring to delete item", 0, nil, w)
			return
		}
		s.Respond(r.Context(), http.StatusOK, "item deleted", 0, nil, w)
	}
}
