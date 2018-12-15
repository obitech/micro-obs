package item

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
		{"GET", "/items", http.StatusOK},
		{"POST", "/items", http.StatusUnprocessableEntity},
		{"PUT", "/items", http.StatusUnprocessableEntity},
		{"DELETE", "/", http.StatusMethodNotAllowed},
		{"DELETE", "/items", http.StatusMethodNotAllowed},
	}
)

func helperSendJSONItem(item *Item, s *Server, method, path string, want int, t *testing.T) {
	js, err := json.Marshal(item)
	if err != nil {
		t.Errorf("Unable to marshal %#v: %s", item, err)
	}
	req, err := http.NewRequest(method, path, bytes.NewBuffer(js))
	if err != nil {
		t.Errorf("unable to create buffer from %s: %s", js, err)
	}
	req.Header.Set("Content-Tye", "application/json")

	w := httptest.NewRecorder()
	s.serveHTTP(w, req)

	if w.Code != want {
		t.Logf("wrong status code on request %#v %#v. Got: %d, want: %d", method, path, w.Code, want)

		res := w.Result()
		b, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()
		t.Logf("revceived: %s", b)
		t.Fail()
	}
}

func helperSendSimpleRequest(s *Server, method, path string, want int, t *testing.T) {
	body := bytes.NewBuffer([]byte{})
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		t.Errorf("unable to create request %#v %#v : %#v", method, path, err)
	}

	w := httptest.NewRecorder()
	s.serveHTTP(w, req)

	if w.Code != want {
		t.Logf("wrong status code on request %#v %#v. Got: %d, want: %d", method, path, w.Code, want)

		res := w.Result()
		b, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()
		t.Logf("revceived: %s", b)
		t.Fail()
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

		t.Run("Individual items", func(t *testing.T) {
			_, mr := helperPrepareMiniredis(t)
			defer mr.Close()

			s, err := NewServer(
				SetRedisAddress(strings.Join([]string{"redis://", mr.Addr()}, "")),
			)
			if err != nil {
				t.Errorf("unable to create server: %s", err)
			}

			t.Run("GET all items", func(t *testing.T) {
				method = "GET"
				want = http.StatusOK
				helperSendSimpleRequest(s, method, path, want, t)
			})

			t.Run("POST new item", func(t *testing.T) {
				method = "POST"
				want = http.StatusCreated

				i := sampleItems[0]
				item, _ := NewItem(i.name, i.desc, i.qty)
				t.Run("Create item in redis", func(t *testing.T) {
					helperSendJSONItem(item, s, method, path, want, t)
				})
				t.Run("Verify item existence", func(t *testing.T) {
					check, err := s.GetItem(item.ID)
					if err != nil {
						t.Errorf("unable to get item with key %s: %s", i.name, err)
					}
					if !reflect.DeepEqual(item, check) {
						t.Errorf("%#v != %#v", item, check)
					}
				})
			})

			t.Run("POST existing item", func(t *testing.T) {
				i := sampleItems[0]
				item, _ := NewItem(i.name, i.desc, i.qty)
				t.Run("Create item in redis", func(t *testing.T) {
					want = http.StatusOK
					helperSendJSONItem(item, s, method, path, want, t)
				})
				t.Run("Verify item existence", func(t *testing.T) {
					check, err := s.GetItem(item.ID)
					if err != nil {
						t.Errorf("unable to get item with key %s: %s", i.name, err)
					}
					if !reflect.DeepEqual(item, check) {
						t.Errorf("%#v != %#v", item, check)
					}
				})
			})

			t.Run("PUT new item", func(t *testing.T) {
				i := sampleItems[1]
				item, _ := NewItem(i.name, i.desc, i.qty)
				t.Run("Create item in redis", func(t *testing.T) {
					method = "PUT"
					want = http.StatusOK
					helperSendJSONItem(item, s, method, path, want, t)
				})
				t.Run("Verify item existence", func(t *testing.T) {
					check, err := s.GetItem(item.ID)
					if err != nil {
						t.Errorf("unable to get item with key %s: %s", i.name, err)
					}
					if !reflect.DeepEqual(item, check) {
						t.Errorf("%#v != %#v", item, check)
					}
				})
			})

			t.Run("PUT existing item", func(t *testing.T) {
				i := sampleItems[1]
				item, _ := NewItem(i.name, i.desc, i.qty)
				t.Run("Create item in redis", func(t *testing.T) {
					method = "PUT"
					want = http.StatusOK
					helperSendJSONItem(item, s, method, path, want, t)
				})
				t.Run("Verify item existence", func(t *testing.T) {
					check, err := s.GetItem(item.ID)
					if err != nil {
						t.Errorf("unable to get item with key %s: %s", i.name, err)
					}
					if !reflect.DeepEqual(item, check) {
						t.Errorf("%#v != %#v", item, check)
					}
				})
			})

			t.Run("GET all items", func(t *testing.T) {
				method = "GET"
				want = http.StatusOK
				helperSendSimpleRequest(s, method, path, want, t)
			})

			t.Run("DELETE single item -> all items", func(t *testing.T) {
				method = "DELETE"
				want = http.StatusOK

				for _, v := range sampleItems {
					item, _ := NewItem(v.name, v.desc, v.qty)
					path = fmt.Sprintf("/items/%s", item.ID)
					helperSendSimpleRequest(s, method, path, want, t)
				}
			})
		})

		t.Run("POST all items", func(t *testing.T) {
			_, mr := helperPrepareMiniredis(t)
			defer mr.Close()

			s, err := NewServer(
				SetRedisAddress(strings.Join([]string{"redis://", mr.Addr()}, "")),
			)
			if err != nil {
				t.Errorf("unable to create server: %s", err)
			}

			path = "/items"

			t.Run("GET all items", func(t *testing.T) {
				method = "GET"
				want = http.StatusOK
				helperSendSimpleRequest(s, method, path, want, t)
			})

			t.Run("POST new item", func(t *testing.T) {
				method = "POST"
				want = http.StatusCreated

				for _, v := range sampleItems {
					item, _ := NewItem(v.name, v.desc, v.qty)

					t.Run("Create item in redis", func(t *testing.T) {
						helperSendJSONItem(item, s, method, path, want, t)
					})

					t.Run("Verify item existence", func(t *testing.T) {
						check, err := s.GetItem(item.ID)
						if err != nil {
							t.Errorf("unable to get item with key %s: %s", item.ID, err)
						}
						if !reflect.DeepEqual(item, check) {
							t.Errorf("%#v != %#v", item, check)
						}
					})
				}
			})

			t.Run("GET all items", func(t *testing.T) {
				method = "GET"
				want = http.StatusOK
				helperSendSimpleRequest(s, method, path, want, t)
			})

			t.Run("GET single item", func(t *testing.T) {
				method = "GET"
				want = http.StatusOK
				for _, v := range sampleItems {
					item, _ := NewItem(v.name, v.desc, v.qty)
					path = fmt.Sprintf("/items/%s", item.ID)
					helperSendJSONItem(item, s, method, path, want, t)
				}
			})

			t.Run("DELTE single item", func(t *testing.T) {
				method = "DELETE"
				want = http.StatusOK
				for _, v := range sampleItems {
					item, _ := NewItem(v.name, v.desc, v.qty)
					path = fmt.Sprintf("/items/%s", item.ID)
					helperSendJSONItem(item, s, method, path, want, t)
				}
			})
		})

		t.Run("PUT all items", func(t *testing.T) {
			_, mr := helperPrepareMiniredis(t)
			defer mr.Close()

			s, err := NewServer(
				SetRedisAddress(strings.Join([]string{"redis://", mr.Addr()}, "")),
			)
			if err != nil {
				t.Errorf("unable to create server: %s", err)
			}

			path = "/items"
			method = "PUT"
			want = http.StatusOK

			t.Run("PUT new item", func(t *testing.T) {
				for _, v := range sampleItems {
					item, _ := NewItem(v.name, v.desc, v.qty)

					t.Run("Create item in redis", func(t *testing.T) {
						helperSendJSONItem(item, s, method, path, want, t)
					})

					t.Run("Verify item existence", func(t *testing.T) {
						check, err := s.GetItem(item.ID)
						if err != nil {
							t.Errorf("unable to get item with key %s: %s", item.ID, err)
						}
						if !reflect.DeepEqual(item, check) {
							t.Errorf("%#v != %#v", item, check)
						}
					})
				}
			})

			t.Run("GET all items", func(t *testing.T) {
				method = "GET"
				want = http.StatusOK
				helperSendSimpleRequest(s, method, path, want, t)
			})

			t.Run("GET single item", func(t *testing.T) {
				method = "GET"
				want = http.StatusOK
				for _, v := range sampleItems {
					item, _ := NewItem(v.name, v.desc, v.qty)
					path = fmt.Sprintf("/items/%s", item.ID)
					helperSendJSONItem(item, s, method, path, want, t)
				}
			})

			t.Run("DELTE single item", func(t *testing.T) {
				method = "DELETE"
				want = http.StatusOK
				for _, v := range sampleItems {
					item, _ := NewItem(v.name, v.desc, v.qty)
					path = fmt.Sprintf("/items/%s", item.ID)
					helperSendJSONItem(item, s, method, path, want, t)
				}
			})
		})
	})
}
