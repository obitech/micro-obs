package util

import "net/http"

//Route defines a specific route
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

// Routes defines a slice of all available API Routes
type Routes []Route
