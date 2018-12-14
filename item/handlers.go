package item

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

// pong sends a simple JSON response.
func (s *Server) pong() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.Respond(http.StatusOK, "pong", 0, nil, w)
	}
}

// getAllItems retrieves all items from Redis.
func (s *Server) getAllItems() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defaultErrMsg := "unable to retrieve items"

		keys, err := s.ScanKeys()
		if err != nil {
			s.logger.Errorw("unable to SCAN redis for keys",
				"error", err,
			)
			s.Respond(http.StatusInternalServerError, defaultErrMsg, 0, nil, w)
			return
		}

		var items = []*Item{}
		for _, k := range keys {
			i, err := s.GetItem(k)
			if err != nil {
				s.logger.Errorw("unable to retrieve item",
					"key", k,
					"error", err,
				)
				s.Respond(http.StatusInternalServerError, defaultErrMsg, 0, nil, w)
				return
			}
			items = append(items, i)
		}

		l := len(items)
		if l == 0 {
			s.Respond(http.StatusOK, "no items present", 0, nil, w)
			return
		}

		s.Respond(http.StatusOK, "items retrieved", l, items, w)
	}
}

// setItem sets an Item as a hash in Redis with URL parameters.
func (s *Server) setItem(update bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			defaultErrMsg string
			item          = &Item{}
		)

		switch update {
		case true:
			defaultErrMsg = "unable to update item"
		case false:
			defaultErrMsg = "unable to create item"
		}

		// Accept payload
		body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
		if err != nil {
			s.logger.Errorw("unable to read request body",
				"error", err,
			)
			s.Respond(http.StatusInternalServerError, "unable to read payload", 0, nil, w)
			return
		}
		defer r.Body.Close()

		// Parse payload
		if err := json.Unmarshal(body, item); err != nil {
			s.logger.Errorw("unable to parse payload",
				"error", err,
			)
			s.Respond(http.StatusUnprocessableEntity, "unable to parse payload", 0, nil, w)
			return
		}
		if err := item.SetID(); err != nil {
			s.logger.Errorw("unable to set item ID",
				"error", err,
			)
			s.Respond(http.StatusInternalServerError, defaultErrMsg, 0, nil, w)
			return
		}
		s.logger.Debugw("item received",
			"id", item.ID,
			"name", item.Name,
			"desc", item.Desc,
			"qty", item.Qty,
		)

		// Check for key existence
		i, err := s.GetItem(item.ID)
		if err != nil {
			s.logger.Errorw("unable to retrieve Item from Redis",
				"key", item.ID,
				"error", err,
			)
			s.Respond(http.StatusInternalServerError, defaultErrMsg, 0, nil, w)
			return
		}
		if i != nil {
			if !update {
				s.Respond(http.StatusBadRequest, fmt.Sprintf("item with name %s already exists", item.Name), 0, nil, w)
				return
			}
		}

		// Create Item in Redis
		err = s.SetItem(item)
		if err != nil {
			s.logger.Errorw("unable to create Item in redis",
				"key", item.ID,
				"error", err,
			)
			s.Respond(http.StatusInternalServerError, defaultErrMsg, 0, nil, w)
			return
		}

		s.Respond(http.StatusCreated, fmt.Sprintf("item %s created", item.Name), 1, []*Item{item}, w)
	}
}
