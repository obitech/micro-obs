package util

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	validResponses = []struct {
		status  int
		message string
		count   int
		data    interface{}
	}{
		{200, "test", 0, nil},
		{404, "not found", 0, nil},
		{200, "int", 1, 1},
		{200, "float", 1, .42},
		{200, "rune", 1, 'a'},
		{200, "slice of ints", 3, []int{1, 2, 3}},
		{200, "slice of floats", 3, []float64{.42, .35234, .005}},
		{200, "slice of strings", 2, []string{"hello", "world"}},
		{200, "slice of runes", 2, []rune{'r', 'u'}},
	}

	responses = []Response{}
)

func helperSendJSON(rd Response, t *testing.T) {
	// Start HTTP Test Server
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := rd.SendJSON(w)
		if err != nil {
			t.Errorf("error sending JSON: %#v", err)
		}
	}))
	defer s.Close()

	// Send request
	res, err := http.Get(s.URL)
	if err != nil {
		t.Errorf("unable to GET response: %#v", err)
	}

	// Retrieve data from response
	d, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Errorf("unable to read response body from request %#v", rd)
	}

	// Verify response
	var v Response
	err = json.Unmarshal(d, &v)
	if err != nil {
		t.Errorf("unable to unmarshal response %#v", d)
	}
}

func TestNewResponse(t *testing.T) {
	for _, tt := range validResponses {
		r, err := NewResponse(tt.status, tt.message, tt.count, tt.data)
		if err != nil {
			t.Errorf("unable to create Response: %#v", err)
		}
		switch {
		case r.Status != tt.status:
			t.Errorf("verifying Response, got: %#v, want: %#v", r.Status, tt.status)
		case r.Message != tt.message:
			t.Errorf("verifying Response, got: %#v, want: %#v", r.Message, tt.message)
		case r.Count != tt.count:
			t.Errorf("verifying Response, got: %#v, want: %#v", r.Count, tt.count)
			// case r.Data != tt.data:
			// 	t.Errorf("verifying Response, got: %#v, want: %#v", r.Data, tt.data)
		}

		responses = append(responses, r)
	}
}

// TODO: figure out an efficient way to table-test this
func TestSendJSON(t *testing.T) {
	for _, rd := range responses {
		helperSendJSON(rd, t)
	}
}
