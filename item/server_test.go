package item

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

var validListeningAddresses = []string{
	"127.0.0.1:8080",
	":80",
	":8080",
	"127.0.0.1:80",
	"192.0.2.1:http",
}

var logLevels = []struct {
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

var validEndpointAddresses = append(validListeningAddresses, []string{
	"golang.org:80",
	"golang.org:http",
}...)

var invalidListeningAddresses = []string{
	":9999999",
	":-1",
	"asokdklasd",
	"0.0.0.0.0.0:80",
	"256.0.0.1:80",
}

var validEndpoints = []struct {
	method     string
	path       string
	wantStatus int
}{
	{"GET", "/", 200},
	{"GET", "/healthz", 200},
	{"GET", "/asdasd", 404},
	{"GET", "/metrics", 200},
}

func TestNewServer(t *testing.T) {
	t.Run("Creating new default server", func(t *testing.T) {
		if _, err := NewServer(); err != nil {
			t.Errorf("error while creating new item server: %s", err)
		}
	})

	t.Run("Creating new default server with different log levels", func(t *testing.T) {
		for _, tt := range logLevels {
			if _, err := NewServer(SetLogLevel(tt.level)); tt.want != nil {
				t.Errorf("error while creating new item server: %s", err)
			}
		}
	})

	t.Run("Creating new item server with valid addresses", func(t *testing.T) {
		for _, listen := range validListeningAddresses {
			for _, ep := range validEndpointAddresses {
				_, err := NewServer(
					SetServerAddress(listen),
					SetServerEndpoint(ep),
				)
				if err != nil {
					t.Errorf("error while creating new item server: %s", err)
				}
			}
		}
	})

	t.Run("Creating new server with invalid addresses", func(t *testing.T) {
		for _, tt := range invalidListeningAddresses {
			_, err := NewServer(
				SetServerAddress(tt),
				SetServerEndpoint(tt),
			)
			if err == nil {
				t.Errorf("Expected error when creating item server with listening address %s, got %s", tt, err)
			}
		}
	})
}

func TestEndpoints(t *testing.T) {
	t.Run("Testing endpoints", func(t *testing.T) {
		s, _ := NewServer()

		for _, tt := range validEndpoints {
			req, err := http.NewRequest(tt.method, tt.path, nil)
			if err != nil {
				t.Errorf("Error creating request %s %s : %s", tt.method, tt.path, err)
			}

			w := httptest.NewRecorder()
			s.ServeHTTP(w, req)
			if w.Code != tt.wantStatus {
				t.Errorf("Wrong status code on request %s %s. Got: %d, want: %d", tt.method, tt.path, w.Code, tt.wantStatus)
			}
		}
	})
}
