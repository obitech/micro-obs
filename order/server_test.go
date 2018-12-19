package order

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/obitech/micro-obs/util"
)

var (
	validListeningAddr = []string{
		"127.0.0.1:8080",
		":80",
		":8080",
		"127.0.0.1:80",
		"192.0.2.1:http",
	}

	invalidListeningAddr = []string{
		":9999999",
		":-1",
		"asokdklasd",
		"0.0.0.0.0.0:80",
		"256.0.0.1:80",
	}

	logLevels = []struct {
		level string
		want  error
	}{
		{"warn", nil},
		{"info", nil},
		{"debug", nil},
		{"error", nil},
		{"", nil},
		{"asdo1293", nil},
		{"😍", nil},
		{"👾 🙇 💁 🙅 🙆 🙋 🙎 🙍", nil},
		{"﷽", nil},
	}

	validRedisAddr = []string{
		"redis://127.0.0.1:6379/0",
		"redis://:qwerty@localhost:6379/1",
		"redis://test:55550",
	}

	invalidRedisAddr = []string{
		"http://localhost:6379",
		"http://google.com",
		"https://google.com",
		"ftp://ftp.fu-berlin.de:21",
	}

	validEndpointAddr = append(validListeningAddr, []string{
		"golang.org:80",
		"golang.org:http",
	}...)

	basicEndpoints = []struct {
		method     string
		path       string
		wantStatus int
	}{
		{"GET", "/", http.StatusOK},
		{"GET", "/healthz", http.StatusOK},
		{"GET", "/asdasd", http.StatusNotFound},
		{"GET", "/metrics", http.StatusOK},
		{"GET", "/orders", http.StatusNotFound},
		{"POST", "/orders", http.StatusBadRequest},
		{"PUT", "/orders", http.StatusBadRequest},
	}

	validJSON = []string{
		`{"id": 99, "items": [{"id": "aab", "qty": 1000}]}`,
		`{"items": [{"id": "asdyb", "qty": 0}], "id": 98} `,
		`{"id": 2018, "items": [{"id": "aab", "qty": 1000}, {"id": "asdyb", "qty": 0}]}`,
	}

	invalidJSON = []struct {
		js   string
		want int
	}{
		{`test`, http.StatusBadRequest},
		{`{`, http.StatusBadRequest},
		{`😍`, http.StatusBadRequest},
		{`{}`, http.StatusUnprocessableEntity},
		{`{"cat": "dog"}`, http.StatusUnprocessableEntity},
		{`{"name": "test", "age": 5}`, http.StatusUnprocessableEntity},
		{`{"id": 15, "desc": "nope"}`, http.StatusUnprocessableEntity},
	}
)

func helperPrepareRedis(t *testing.T) (*miniredis.Miniredis, *Server) {
	_, mr := helperPrepareMiniredis(t)

	s, err := NewServer(
		SetRedisAddress(strings.Join([]string{"redis://", mr.Addr()}, "")),
	)
	if err != nil {
		t.Errorf("unable to create server: %s", err)
	}

	return mr, s
}

func helperSendJSON(valid bool, js []byte, s *Server, method, path string, want int, t *testing.T) {
	var (
		err       error
		req       *http.Request
		res       *http.Response
		b         []byte
		vResponse Response
	)

	req, err = http.NewRequest(method, path, bytes.NewBuffer(js))
	if err != nil {
		t.Errorf("unable to create buffer from %#v: %s", js, err)
	}
	req.Header.Set("Content-Tye", "application/json")

	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	res = w.Result()
	b, err = ioutil.ReadAll(res.Body)
	if err != nil {
		t.Errorf("unable to read response body: %s", err)
	}
	defer res.Body.Close()

	if w.Code != want {
		t.Logf("wrong status code on request %#v %#v. Got: %d, want: %d", method, path, w.Code, want)
		t.Logf("revceived: %s", b)
		t.Fail()
	}

	err = json.Unmarshal(b, &vResponse)
	if err != nil {
		t.Errorf("unable to unmarshal into response: %s", err)
	}

}

func helperSendJSONOrder(order *Order, s *Server, method, path string, want int, t *testing.T) {
	var (
		js        []byte
		err       error
		req       *http.Request
		res       *http.Response
		b         []byte
		vResponse Response
	)

	js, err = json.Marshal(order)
	if err != nil {
		t.Errorf("unable to marshal %#v: %s", order, err)
	}

	req, err = http.NewRequest(method, path, bytes.NewBuffer(js))
	if err != nil {
		t.Errorf("unable to create buffer from %s: %s", js, err)
	}

	req.Header.Set("Content-Tye", "application/json")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	res = w.Result()
	b, err = ioutil.ReadAll(res.Body)
	if err != nil {
		t.Errorf("unable to read response body: %s", err)
	}
	defer res.Body.Close()

	if w.Code != want {
		t.Logf("wrong status code on request %#v %#v. Got: %d, want: %d", method, path, w.Code, want)
		t.Logf("revceived: %s", b)
		t.Fail()
	}

	err = json.Unmarshal(b, &vResponse)
	if err != nil {
		t.Errorf("unable to unmarshal into response: %s", err)
	}

	for _, vOrder := range vResponse.Data {
		if !reflect.DeepEqual(vOrder, order) {
			t.Errorf("%+v != %+v", vOrder, order)
		}
	}

}

func helperSendSimpleRequest(s util.Server, method, path string, want int, t *testing.T) {
	var (
		err error
		req *http.Request
		res *http.Response
		b   []byte
	)

	body := bytes.NewBuffer([]byte{})
	req, err = http.NewRequest(method, path, body)
	if err != nil {
		t.Errorf("unable to create request %#v %#v : %#v", method, path, err)
	}

	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	res = w.Result()
	b, err = ioutil.ReadAll(res.Body)
	if err != nil {
		t.Errorf("unable to read body: %s", err)
	}
	defer res.Body.Close()

	if w.Code != want {
		t.Logf("wrong status code on request %#v %#v. Got: %d, want: %d", method, path, w.Code, want)
		t.Logf("revceived: %s", b)
		t.Fail()
	}

}

func helperSendJSONandVerify(s util.Server, method, path string, want int, t *testing.T, orders ...*Order) {
	var (
		err       error
		req       *http.Request
		res       *http.Response
		b         []byte
		vResponse Response
	)

	req, err = http.NewRequest(method, path, nil)
	if err != nil {
		t.Errorf("unable to create request %#v %#v : %#v", method, path, err)
	}

	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	res = w.Result()
	b, err = ioutil.ReadAll(res.Body)
	if err != nil {
		t.Errorf("unable to read body: %s", err)
	}
	defer res.Body.Close()

	if w.Code != want {
		t.Logf("wrong status code on request %#v %#v. Got: %d, want: %d", method, path, w.Code, want)
		t.Logf("revceived: %s", b)
		t.Fail()
	}

	err = json.Unmarshal(b, &vResponse)
	if err != nil {
		t.Errorf("unable to unmarshal response: %s", err)
	}

	if len(vResponse.Data) == 0 || vResponse.Data == nil {
		t.Errorf("no items found in %+v", vResponse)
	}

	for i, v := range orders {
		t.Logf("order: %+v", v)
		if !reflect.DeepEqual(v, vResponse.Data[i]) {
			t.Errorf("%+v != %+v", v, vResponse.Data[i])
		}
	}

}

func TestNewServer(t *testing.T) {
	t.Run("Creating new default server", func(t *testing.T) {
		if _, err := NewServer(); err != nil {
			t.Errorf("error while creating new item server: %#v", err)
		}
	})

	t.Run("Creating new default server with custom log levels", func(t *testing.T) {
		for _, tt := range logLevels {
			if _, err := NewServer(SetLogLevel(tt.level)); tt.want != nil {
				t.Errorf("error while creating new item server: %#v", err)
			}
		}
	})

	t.Run("Creating new default server with custom redis address", func(t *testing.T) {
		t.Run("Checking valid addresses", func(t *testing.T) {
			for _, v := range validRedisAddr {
				if _, err := NewServer(SetRedisAddress(v)); err != nil {
					t.Errorf("error while creating new item server: %#v", err)
				}
			}
		})

		t.Run("Checking invalid addresses", func(t *testing.T) {
			for _, v := range invalidRedisAddr {
				if _, err := NewServer(SetRedisAddress(v)); err == nil {
					t.Errorf("expected error while setting redis address to %#v, got %#v", v, err)
				}
			}
		})
	})

	t.Run("Creating new server with custom listening addresses", func(t *testing.T) {
		t.Run("Checking valid addresses", func(t *testing.T) {
			for _, listen := range validListeningAddr {
				for _, ep := range validEndpointAddr {
					_, err := NewServer(
						SetServerAddress(listen),
						SetServerEndpoint(ep),
					)
					if err != nil {
						t.Errorf("error while creating new item server: %#v", err)
					}
				}
			}
		})
		t.Run("Checking invalid addresses", func(t *testing.T) {
			for _, tt := range invalidListeningAddr {
				if _, err := NewServer(
					SetServerAddress(tt),
					SetServerEndpoint(tt),
				); err == nil {
					t.Errorf("expected error when creating item server with listening address %#v, got %#v", tt, err)
				}
			}
		})
	})

	t.Run("Creating new server with custom itemService address", func(t *testing.T) {
		t.Run("Checking valid addresses", func(t *testing.T) {
			for _, v := range validListeningAddr {
				if _, err := NewServer(SetItemServiceAddress(v)); err != nil {
					t.Errorf("error while creating new item server: %#v", err)
				}
			}
		})

		t.Run("Checking invalid addresses", func(t *testing.T) {
			for _, v := range invalidListeningAddr {
				if _, err := NewServer(SetItemServiceAddress(v)); err == nil {
					t.Errorf("expected error while setting redis address to %#v, got %#v", v, err)
				}
			}
		})
	})
}

func TestEndpoints(t *testing.T) {
	t.Run("Basic endpoints", func(t *testing.T) {
		_, mr := helperPrepareMiniredis(t)
		defer mr.Close()

		s, err := NewServer(
			SetRedisAddress(strings.Join([]string{"redis://", mr.Addr()}, "")),
		)
		if err != nil {
			t.Errorf("unable to create server: %s", err)
		}

		for _, tt := range basicEndpoints {
			helperSendSimpleRequest(s, tt.method, tt.path, tt.wantStatus, t)
		}
	})

	t.Run("Orders Endpoint", func(t *testing.T) {
		var (
			path   = "/orders"
			method = "GET"
			want   = http.StatusNotFound
		)

		t.Run("POST orders", func(t *testing.T) {
			mr, s := helperPrepareRedis(t)
			defer mr.Close()

			t.Run("GET all empty orders", func(t *testing.T) {
				helperSendSimpleRequest(s, method, path, want, t)
			})

			t.Run("New orders", func(t *testing.T) {
				t.Run("From Structs", func(t *testing.T) {
					method = "POST"
					want = http.StatusCreated
					path = "/orders"

					for _, o := range uniqueOrders {
						helperSendJSONOrder(o, s, method, path, want, t)
					}
					
					t.Run("Verifying with GET", func(t *testing.T) {
						for _, o := range uniqueOrders {
								method = "GET"
								want = http.StatusOK
								path = fmt.Sprintf("/orders/%d", o.ID)

								helperSendJSONandVerify(s, method, path, want, t, o)
						}
					})
				})

				t.Run("From raw JSON", func(t *testing.T) {
					method = "POST"
					want = http.StatusCreated
					path = "/orders"

					for _, js := range validJSON {
						helperSendJSON(true, []byte(js), s, method, path, want, t)
					}
				})

				t.Run("Invalid JSON", func(t *testing.T) {
					method = "POST"
					path = "/orders"

					for _, tt := range invalidJSON {
						helperSendJSON(false, []byte(tt.js), s, method, path, tt.want, t)
					}
				})
			})

			t.Run("GET all filled orders", func(t *testing.T) {
				method = "GET"
				want = http.StatusOK
				path = "/orders"

				helperSendSimpleRequest(s, method, path, want, t)
			})

			t.Run("Existing orders", func(t *testing.T) {
				method = "POST"
				want = http.StatusUnprocessableEntity
				path = "/orders"

				t.Run("From Structs", func(t *testing.T) {
					for _, o := range uniqueOrders {
						helperSendJSONOrder(o, s, method, path, want, t)
					}
				})

				t.Run("From raw JSON", func(t *testing.T) {
					for _, js := range validJSON {
						helperSendJSON(true, []byte(js), s, method, path, want, t)
					}
				})
			})
		})

		t.Run("PUT orders", func(t *testing.T) {
			mr, s := helperPrepareRedis(t)
			defer mr.Close()

			t.Run("New orders", func(t *testing.T) {
				t.Run("From Structs", func(t *testing.T) {
					method = "PUT"
					want = http.StatusOK
					path = "/orders"

					for _, o := range uniqueOrders {
						helperSendJSONOrder(o, s, method, path, want, t)
					}
				})

				t.Run("From raw JSON", func(t *testing.T) {
					method = "PUT"
					want = http.StatusOK
					path = "/orders"

					for _, js := range validJSON {
						helperSendJSON(true, []byte(js), s, method, path, want, t)
					}
				})

				t.Run("Invalid JSON", func(t *testing.T) {
					method = "PUT"
					want = http.StatusOK
					path = "/orders"

					for _, tt := range invalidJSON {
						helperSendJSON(false, []byte(tt.js), s, method, path, tt.want, t)
					}
				})
			})

			t.Run("Existing orders", func(t *testing.T) {
				t.Run("From Structs", func(t *testing.T) {
					method = "PUT"
					want = http.StatusOK
					path = "/orders"

					for _, o := range uniqueOrders {
						helperSendJSONOrder(o, s, method, path, want, t)
					}
				})

				t.Run("From raw JSON", func(t *testing.T) {
					method = "PUT"
					want = http.StatusOK
					path = "/orders"

					for _, js := range validJSON {
						helperSendJSON(true, []byte(js), s, method, path, want, t)
					}
				})
			})
		})
	})

}
