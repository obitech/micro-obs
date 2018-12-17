package util

import "net/http"

// Server is an interface for types that implement ServeHTTP.
// For now this is only used for testing but will possibly extended for refactoring at a later point.
type Server interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}
