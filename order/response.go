package order

import (
	"encoding/json"
	"net/http"
)

// Response defines an API response.
type Response struct {
	Status  int      `json:"status"`
	Message string   `json:"message"`
	Count   int      `json:"count"`
	Data    []*Order `json:"data"`
}

// NewResponse returns a Response with a passed message string and slice of Data.
// TODO: make d variadic
func NewResponse(s int, m string, c int, d []*Order) (Response, error) {
	return Response{
		Status:  s,
		Message: m,
		Count:   c,
		Data:    d,
	}, nil
}

// SendJSON encodes a Response as JSON and sends it on a passed http.ResponseWriter.
func (r Response) SendJSON(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/JSON; charset=UTF-8")
	w.WriteHeader(r.Status)
	err := json.NewEncoder(w).Encode(r)
	return err
}
