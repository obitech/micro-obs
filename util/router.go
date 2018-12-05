package util

import "github.com/gorilla/mux"

// NewRouter returns a gorilla/mux router
func NewRouter() *mux.Router {
	router := mux.NewRouter()
	return router
}
