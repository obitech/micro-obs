package item

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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
		{"GET", "/", 200},
		{"GET", "/healthz", 200},
		{"GET", "/asdasd", 404},
		{"GET", "/metrics", 200},
		{"GET", "/items", 200},
		{"POST", "/items", 422},
		{"PUT", "/items", 422},
	}
)

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
					t.Errorf("Expected error while setting redis address to %#v, got %#v", v, err)
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
					t.Errorf("Expected error when creating item server with listening address %#v, got %#v", tt, err)
				}
			}
		})
	})
}

func TestEndpoints(t *testing.T) {
	_, mr := helperPrepareMiniredis(t)
	defer mr.Close()

	s, err := NewServer(
		SetRedisAddress(strings.Join([]string{"redis://", mr.Addr()}, "")),
	)
	if err != nil {
		t.Errorf("unable to create server: %s", err)
	}

	t.Run("Checking for 200", func(t *testing.T) {
		for _, tt := range basicEndpoints {
			body := bytes.NewBuffer([]byte{})
			req, err := http.NewRequest(tt.method, tt.path, body)
			if err != nil {
				t.Errorf("Error creating request %#v %#v : %#v", tt.method, tt.path, err)
			}

			w := httptest.NewRecorder()
			s.serveHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Logf("Wrong status code on request %#v %#v. Got: %d, want: %d", tt.method, tt.path, w.Code, tt.wantStatus)

				res := w.Result()
				b, _ := ioutil.ReadAll(res.Body)
				res.Body.Close()
				t.Logf("Revceived: %s", b)
				t.Fail()
			}
		}
	})
}
