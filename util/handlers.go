package util

import (
	"fmt"
	"net/http"
)

// Healthz responds to a HTTP healthcheck
func Healthz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK\n")
	}
}
