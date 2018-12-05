package util

import (
	"encoding/json"
	"net/http"
)

// Response defines an API response.
type Response struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Count   int         `json:"count"`
	Data    interface{} `json:"data"`
}

// NewResponse returns a Response with a passed message string and slice of Data.
// This will automatically set the Status field to 200.
func NewResponse(s int, m string, c int, d interface{}) Response {
	return Response{
		Status:  s,
		Message: m,
		Count:   c,
		Data:    d,
	}
}

// SendJSON encodes a Response as JSON and sends it on a passed http.ResponseWriter.
func (r Response) SendJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/JSON; charset=UTF-8")
	w.WriteHeader(r.Status)
	if err := json.NewEncoder(w).Encode(r); err != nil {
		status := http.StatusInternalServerError
		w.WriteHeader(status)

		r.Status = status
		r.Message = "Internal Server Error while trying to encode JSON"
		r.Data = nil

		_ = json.NewEncoder(w).Encode(r)
	}
}

// SendErrorJSON uses SendJSON to send a passed error code and message
func SendErrorJSON(w http.ResponseWriter, s int, m string) {
	NewResponse(s, m, 0, nil).SendJSON(w)
}
