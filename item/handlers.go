package item

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/obitech/micro-obs/util"
	ot "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
)

// NotFound is a 404 Message according to the Response type
func (s *Server) notFound() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		span, ctx := ot.StartSpanFromContext(r.Context(), "notFound")
		defer span.Finish()
		s.Respond(ctx, http.StatusNotFound, "resource not found", 0, nil, w)
	}
}

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
		log := util.RequestIDLogger(s.logger, r)

		defaultErrMsg := "unable to retrieve items"
		keys, err := s.RedisScanKeys(ctx)
		if err != nil {
			log.Errorw("unable to SCAN redis for keys",
				"error", err,
			)
			s.Respond(ctx, http.StatusInternalServerError, defaultErrMsg, 0, nil, w)
			return
		}

		var items = []*Item{}
		for _, k := range keys {
			i, err := s.RedisGetItem(ctx, k)
			if err != nil {
				log.Errorw("unable to retrieve item",
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
		log := util.RequestIDLogger(s.logger, r)

		var (
			defaultErrMsg = "unable to create items"
			items         []*Item
		)

		// Accept payload
		body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
		if err != nil {
			log.Errorw("unable to read request body",
				"error", err,
			)
			r.Body.Close()
			s.Respond(ctx, http.StatusInternalServerError, "unable to read payload", 0, nil, w)
			return
		}
		defer r.Body.Close()

		// Parse payload
		if err := json.Unmarshal(body, &items); err != nil {
			log.Errorw("unable to parse payload",
				"error", err,
			)
			s.Respond(ctx, http.StatusBadRequest, "unable to parse payload", 0, nil, w)
			return
		}

		if items == nil || len(items) == 0 {
			s.Respond(ctx, http.StatusUnprocessableEntity, "items can't be empty", 0, nil, w)
		}

		// Verify sent items
		for _, item := range items {
			// Catch empty response
			if item.Name == "" {
				s.Respond(ctx, http.StatusUnprocessableEntity, "item needs name", 0, nil, w)
				return
			}

			if err := item.SetID(ctx); err != nil {
				log.Errorw("unable to set item ID",
					"error", err,
				)
				s.Respond(ctx, http.StatusInternalServerError, defaultErrMsg, 0, nil, w)
				return
			}

			log.Debugw("item struct created",
				"id", item.ID,
				"name", item.Name,
				"desc", item.Desc,
				"qty", item.Qty,
			)

			// Check for existence
			_, err := s.RedisGetItem(ctx, item.ID)
			if err != nil {
				log.Errorw("unable to retrieve Item from Redis",
					"key", item.ID,
					"error", err,
				)
				s.Respond(ctx, http.StatusInternalServerError, defaultErrMsg, 0, nil, w)
				return
			}
		}

		var itemsCreatedMsg string
		var itemsCreatedData []*Item
		var itemsFailedMsg string

		for _, item := range items {
			// Check for existence
			i, err := s.RedisGetItem(ctx, item.ID)
			if err != nil {
				log.Errorw("unable to retrieve Item from Redis",
					"key", item.ID,
					"error", err,
				)
				s.Respond(ctx, http.StatusInternalServerError, "unable to create items", 0, nil, w)
				return
			}
			if i != nil {
				if !update {
					log.Debugw("item already exists",
						"key", item.ID,
					)
					itemsFailedMsg += fmt.Sprintf("%s already exists, ", item.ID)
					continue
				}
			}

			// Create Item in Redis
			err = s.RedisSetItem(ctx, item)
			if err != nil {
				log.Errorw("unable to create item in redis",
					"key", item.ID,
					"error", err,
				)
				itemsFailedMsg += fmt.Sprintf("%s, ", item.ID)
				continue
			}
			itemsCreatedMsg += fmt.Sprintf("%s, ", item.ID)
			itemsCreatedData = append(itemsCreatedData, item)
		}

		switch {
		// All failed
		case len(itemsFailedMsg) != 0 && len(itemsCreatedData) == 0:
			msg := fmt.Sprintf("unable to create items: %s", itemsFailedMsg)
			s.Respond(ctx, http.StatusUnprocessableEntity, msg, 0, nil, w)
			return

		// Some created, some failed
		case len(itemsFailedMsg) != 0 && len(itemsCreatedData) > 0:
			msg := fmt.Sprintf("items %s created but some failed: %s", itemsCreatedMsg, itemsFailedMsg)
			s.Respond(ctx, http.StatusOK, msg, len(itemsCreatedData), itemsCreatedData, w)
			return
		}

		// All created, none failed
		s.Respond(ctx, http.StatusCreated, fmt.Sprintf("items %s created", itemsCreatedMsg), len(itemsCreatedData), itemsCreatedData, w)
	}
}

// getItem retrieves a single Item by ID from Redis.
func (s *Server) getItem() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		span, ctx := ot.StartSpanFromContext(r.Context(), "getItem")
		defer span.Finish()
		log := util.RequestIDLogger(s.logger, r)

		pr := mux.Vars(r)
		key := pr["id"]

		item, err := s.RedisGetItem(ctx, key)
		if err != nil {
			log.Errorw("unable to get key from redis",
				"key", key,
				"error", err,
			)
			s.Respond(ctx, http.StatusInternalServerError, "unable to retreive item", 0, nil, w)
			return
		}
		if item == nil {
			s.Respond(ctx, http.StatusNotFound, fmt.Sprintf("item with ID %s doesn't exist", key), 0, nil, w)
			return
		}
		s.Respond(ctx, http.StatusOK, "item retrieved", 1, []*Item{item}, w)
	}
}

// delItem deletes a single item by ID.
func (s *Server) delItem() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		span, ctx := ot.StartSpanFromContext(r.Context(), "delItem")
		defer span.Finish()
		log := util.RequestIDLogger(s.logger, r)

		pr := mux.Vars(r)
		key := pr["id"]

		err := s.RedisDelItem(ctx, key)
		if err != nil {
			log.Errorw("unable to delete key from redis",
				"key", key,
				"error", err,
			)
			s.Respond(ctx, http.StatusInternalServerError, "an error occured while tring to delete item", 0, nil, w)
			return
		}
		s.Respond(ctx, http.StatusOK, "item deleted", 0, nil, w)
	}
}

// delay returns after a random period to simulate reequest delay.
func (s *Server) delay() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		span, ctx := ot.StartSpanFromContext(r.Context(), "delay")
		defer span.Finish()
		log := util.RequestIDLogger(s.logger, r)

		// Simulate delay between 10ms - 500ms
		t := time.Duration((rand.Float64()*500)+10) * time.Millisecond

		span.SetTag("wait", t)
		log.Debugw("Waiting",
			"time", t,
		)

		time.Sleep(t)

		s.Respond(ctx, http.StatusOK, fmt.Sprintf("waited %v", t), 0, nil, w)
	}
}

// simulateError returns a 500
func (s *Server) simulateError() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		span, ctx := ot.StartSpanFromContext(r.Context(), "simulateError")
		defer span.Finish()
		log := util.RequestIDLogger(s.logger, r)

		log.Errorw("Error occured",
			"error", errors.New("nasty error message"),
		)

		s.Respond(ctx, http.StatusInternalServerError, "simulated error", 0, nil, w)
	}
}
