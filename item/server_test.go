package item

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/alicebob/miniredis"
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
		{"GET", "/items", http.StatusNotFound},
		{"GET", "/delay", http.StatusOK},
		{"POST", "/items", http.StatusBadRequest},
		{"PUT", "/items", http.StatusBadRequest},
		{"DELETE", "/", http.StatusMethodNotAllowed},
		{"DELETE", "/items", http.StatusMethodNotAllowed},
		{"GET", "/error", http.StatusInternalServerError},
	}

	validJSON = []string{

		`[{"name": "😍aa", "qty": 42, "desc": "yes"}]`,
		`[{"name": "😍aabasd", "qty": 42, "desc": "yes"}, {"name": "bread", "desc": "love it", "qty": 15}]`,
	}

	invalidJSON = []struct {
		js   string
		want int
	}{
		{`test`, http.StatusBadRequest},
		{`{}`, http.StatusBadRequest},
		{`{"cat": "dog"}`, http.StatusBadRequest},
		{`[`, http.StatusBadRequest},
		{`{"name": "orange", "desc": "test", "qty": 1}`, http.StatusBadRequest},
		{`{"name": "😍", "qty": 42, "desc": "yes"}`, http.StatusBadRequest},
		{`[{}]`, http.StatusUnprocessableEntity},
		{`[]`, http.StatusUnprocessableEntity},
	}
)

func helperSendJSON(js string, s *Server, method, path string, want int, t *testing.T) []byte {
	req, err := http.NewRequest(method, path, bytes.NewBuffer([]byte(js)))
	if err != nil {
		t.Errorf("unable to create buffer from %#v: %s", js, err)
	}
	req.Header.Set("Content-Tye", "application/json")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	res := w.Result()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Errorf("unable to read response body: %s", err)
	}
	res.Body.Close()

	if w.Code != want {
		t.Logf("%s %s -> %s", method, path, string(js))
		t.Logf("wrong status code on request %#v %#v. Got: %d, want: %d", method, path, w.Code, want)
		t.Logf("revceived: %s", b)
		t.Fail()
	}

	return b
}

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

func helperSendSimpleRequest(s *Server, method, path string, want int, t *testing.T) []byte {
	body := bytes.NewBuffer([]byte{})
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		t.Errorf("unable to create request %#v %#v : %#v", method, path, err)
	}

	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	res := w.Result()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Errorf("unable to read body: %s", err)
	}
	res.Body.Close()

	if w.Code != want {
		t.Logf("wrong status code on request %#v %#v. Got: %d, want: %d", method, path, w.Code, want)
		t.Logf("revceived: %s", b)
		t.Fail()
	}

	return b
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

	t.Run("Creating new item server with custom listening addresses", func(t *testing.T) {
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

	t.Run("Items Endpoint", func(t *testing.T) {
		var (
			path   = "/items"
			method = "POST"
			want   = http.StatusCreated
		)

		t.Run("POST GET DELETE", func(t *testing.T) {
			mr, s := helperPrepareRedis(t)
			defer mr.Close()

			t.Run("GET all empty items", func(t *testing.T) {
				method = "GET"
				want = http.StatusNotFound

				helperSendSimpleRequest(s, method, path, want, t)
			})

			t.Run("POST new item", func(t *testing.T) {
				method = "POST"
				want = http.StatusCreated

				for _, js := range validJSON {
					helperSendJSON(js, s, method, path, want, t)
				}
			})

			t.Run("GET single item", func(t *testing.T) {
				method = "GET"
				want := http.StatusOK

				for _, js := range validJSON {
					// JSON to Item
					var items []*Item
					err := json.Unmarshal([]byte(js), &items)
					if err != nil {
						t.Errorf("unable to unmarshal %+v into item: %s", js, err)
					}

					for _, item := range items {
						err = item.SetID(context.Background())
						if err != nil {
							t.Errorf("unable to set HashID on %+v: %s", item, err)
						}

						path := fmt.Sprintf("/items/%s", item.ID)

						// Send JSON
						b := helperSendSimpleRequest(s, method, path, want, t)
						var verify Response
						err = json.Unmarshal(b, &verify)
						if err != nil {
							t.Errorf("unable to parse response: %s", err)
						}

						// Verify response with item
						if !reflect.DeepEqual(verify.Data, []*Item{item}) {
							t.Errorf("%+v != %+v", verify.Data, []*Item{item})
						}
					}
				}
			})

			t.Run("POST existing item", func(t *testing.T) {
				method = "POST"
				want := http.StatusUnprocessableEntity
				path := "/items"

				for _, js := range validJSON {
					helperSendJSON(js, s, method, path, want, t)
				}
			})

			t.Run("POST invalid JSON", func(t *testing.T) {
				method = "POST"
				want = http.StatusBadRequest

				for _, tt := range invalidJSON {
					helperSendJSON(tt.js, s, method, path, tt.want, t)
				}
			})

			t.Run("GET all items", func(t *testing.T) {
				method = "GET"
				want = http.StatusOK

				b := helperSendSimpleRequest(s, method, path, want, t)

				var verify Response
				err := json.Unmarshal(b, &verify)
				if err != nil {
					t.Errorf("unable to parse response: %s", err)
				}
			})

			t.Run("DELETE all items", func(t *testing.T) {
				method = "DELETE"
				want = http.StatusOK

				for _, js := range validJSON {
					// JSON to Item
					var items []*Item
					err := json.Unmarshal([]byte(js), &items)
					if err != nil {
						t.Errorf("unable to unmarshal %+v into item: %s", js, err)
					}

					for _, item := range items {
						err = item.SetID(context.Background())
						if err != nil {
							t.Errorf("unable to set HashID on %+v: %s", item, err)
						}

						path := fmt.Sprintf("/items/%s", item.ID)
						helperSendSimpleRequest(s, method, path, want, t)
					}
				}
			})
		})

		t.Run("PUT GET DELETE", func(t *testing.T) {
			mr, s := helperPrepareRedis(t)
			defer mr.Close()

			t.Run("PUT new item", func(t *testing.T) {
				method = "PUT"
				want := http.StatusCreated

				for _, js := range validJSON {
					helperSendJSON(js, s, method, path, want, t)
				}
			})

			t.Run("PUT existing item", func(t *testing.T) {
				method = "PUT"
				want := http.StatusCreated

				for _, js := range validJSON {
					helperSendJSON(js, s, method, path, want, t)
				}
			})

			t.Run("GET single item", func(t *testing.T) {
				method = "GET"
				want := http.StatusOK

				for _, js := range validJSON {
					// JSON to Item
					var items []*Item
					err := json.Unmarshal([]byte(js), &items)
					if err != nil {
						t.Errorf("unable to unmarshal %+v into item: %s", js, err)
					}

					for _, item := range items {
						err = item.SetID(context.Background())
						if err != nil {
							t.Errorf("unable to set HashID on %+v: %s", item, err)
						}

						path := fmt.Sprintf("/items/%s", item.ID)

						// Send JSON
						b := helperSendSimpleRequest(s, method, path, want, t)
						var verify Response
						err = json.Unmarshal(b, &verify)
						if err != nil {
							t.Errorf("unable to parse response: %s", err)
						}

						// Verify response with item
						if !reflect.DeepEqual(verify.Data, []*Item{item}) {
							t.Errorf("%+v != %+v", verify.Data, []*Item{item})
						}
					}
				}
			})

			t.Run("PUT invalid JSON", func(t *testing.T) {
				method = "PUT"
				want = http.StatusUnprocessableEntity

				for _, tt := range invalidJSON {
					helperSendJSON(tt.js, s, method, path, tt.want, t)
				}
			})

			t.Run("GET all items", func(t *testing.T) {
				method = "GET"
				want = http.StatusOK

				b := helperSendSimpleRequest(s, method, path, want, t)

				var verify Response
				err := json.Unmarshal(b, &verify)
				if err != nil {
					t.Errorf("unable to parse response: %s", err)
				}
			})

			t.Run("DELETE all items", func(t *testing.T) {
				method = "DELETE"
				want = http.StatusOK

				for _, js := range validJSON {
					// JSON to Item
					var items []*Item
					err := json.Unmarshal([]byte(js), &items)
					if err != nil {
						t.Errorf("unable to unmarshal %+v into item: %s", js, err)
					}

					for _, item := range items {
						err = item.SetID(context.Background())
						if err != nil {
							t.Errorf("unable to set HashID on %+v: %s", item, err)
						}

						path := fmt.Sprintf("/items/%s", item.ID)
						helperSendSimpleRequest(s, method, path, want, t)
					}
				}
			})
		})
	})
}
