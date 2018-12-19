package order

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func helperSendJSONResponse(rd Response, t *testing.T) {
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

func TestResponse(t *testing.T) {
	var validResponses []Response
	t.Run("Test NewResponse", func(t *testing.T) {
		t.Run("nil orders", func(t *testing.T) {
			r, err := NewResponse(200, "test", 0, nil)
			if err != nil {
				t.Errorf("unable to create response: %s", err)
			}
			validResponses = append(validResponses, r)
		})

		t.Run("One order", func(t *testing.T) {
			r, err := NewResponse(200, "test", 1, []*Order{uniqueOrders[0]})
			if err != nil {
				t.Errorf("unable to create response: %s", err)
			}
			validResponses = append(validResponses, r)
		})

		t.Run("Multiple orders", func(t *testing.T) {
			r, err := NewResponse(200, "test", len(uniqueOrders), uniqueOrders)
			if err != nil {
				t.Errorf("unable to create response: %s", err)
			}
			validResponses = append(validResponses, r)
		})
	})

	t.Run("Verify responses", func(t *testing.T) {
		for _, r := range validResponses {
			helperSendJSONResponse(r, t)
		}
	})
}
