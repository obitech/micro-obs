package item

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis"
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
		"redis://test",
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

	validEndpoints = []struct {
		method     string
		path       string
		wantStatus int
	}{
		{"GET", "/", 200},
		{"GET", "/healthz", 200},
		{"GET", "/asdasd", 404},
		{"GET", "/metrics", 200},
	}

	redisSampleKV = []struct {
		key   string
		value string
	}{
		{"foo", "bar"},
		{"hello", "world"},
		{"😍", "test"},
		{"test", "😍"},
		{"👾", "🙅"},
	}

	redisSampleHash = []struct {
		key string
		fv  map[string]string
	}{
		{"hash1", map[string]string{
			"foo": "bar",
		}},
		{"orange", map[string]string{
			"qty":  string("5"),
			"desc": "A jummy fruit",
		}},
	}
)

func redisGET(c *redis.Client, k, want string, t *testing.T) {
	r, err := c.Get(k).Result()
	if err != nil {
		t.Errorf("Unable to GET %s: %s", k, err)
	}
	if r != want {
		t.Errorf("GET %s, expected: %s, got: %s", k, r, want)
	}
}

func redisSET(c *redis.Client, k, v string, t *testing.T) {
	if err := c.Set(k, v, 0).Err(); err != nil {
		t.Errorf("Unable to SET %s -> %s: %s", k, v, err)
	}
}

func redisHSET(c *redis.Client, k, f, v string, want bool, t *testing.T) {
	r, err := c.HSet(k, f, v).Result()
	if err != nil {
		t.Errorf("Unable to HSET %s %s %s: %s", k, f, v, err)
	}
	if r != want {
		t.Errorf("HSET %s %s %s, expected: %t, got: %t", k, f, v, want, r)
	}
}

func redisDEL(c *redis.Client, k string, want int64, t *testing.T) {
	nr, err := c.Del(k).Result()
	if err != nil {
		t.Errorf("Unable to DEL %s: %s", k, err)
	}
	if nr != want {
		t.Errorf("DEL %s, expected: %d, got: %d", k, want, nr)
	}
}

func redisEXISTS(c *redis.Client, k string, want int64, t *testing.T) {
	nr, err := c.Exists(k).Result()
	if err != nil {
		t.Errorf("Unable to EXISTS %s: %s", k, err)
	}
	if nr != want {
		t.Errorf("EXISTS %s, expected: %d, got: %d", k, want, nr)
	}
}

func redisHGETALL(c *redis.Client, k string, want map[string]string, t *testing.T) {
	r, err := c.HGetAll(k).Result()
	if err != nil {
		t.Errorf("Unable to HGETALL %s: %s", k, err)
	}
	for f, v := range want {
		if r[f] != v {
			t.Errorf("HGETALL %s, expected: %s => %s, got: %s => %s", k, f, v, f, r[f])
		}
	}
}

func TestNewServer(t *testing.T) {
	t.Run("Creating new default server", func(t *testing.T) {
		if _, err := NewServer(); err != nil {
			t.Errorf("error while creating new item server: %s", err)
		}
	})

	t.Run("Creating new default server with custom log levels", func(t *testing.T) {
		for _, tt := range logLevels {
			if _, err := NewServer(SetLogLevel(tt.level)); tt.want != nil {
				t.Errorf("error while creating new item server: %s", err)
			}
		}
	})

	t.Run("Creating new default server with custom redis address", func(t *testing.T) {
		t.Run("Checking valid addresses", func(t *testing.T) {
			for _, v := range validRedisAddr {
				if _, err := NewServer(SetRedisAddress(v)); err != nil {
					t.Errorf("error while creating new item server: %s", err)
				}
			}
		})

		t.Run("Checking invalid addresses", func(t *testing.T) {
			for _, v := range invalidRedisAddr {
				if _, err := NewServer(SetRedisAddress(v)); err == nil {
					t.Errorf("Expected error while setting redis address to %s, got %s", v, err)
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
						t.Errorf("error while creating new item server: %s", err)
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
					t.Errorf("Expected error when creating item server with listening address %s, got %s", tt, err)
				}
			}
		})
	})
}

func TestRedis(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Errorf("Unable to create miniredis server: %s", err)
	}
	defer s.Close()

	// Setting sample data
	for _, v := range redisSampleKV {
		if err := s.Set(v.key, v.value); err != nil {
			t.Errorf("Unable to set sample data: %s", err)
		}
	}

	for _, data := range redisSampleHash {
		for f, v := range data.fv {
			s.HSet(data.key, f, v)
		}
	}

	t.Run("Testing go-redis", func(t *testing.T) {
		// Connecting
		c := redis.NewClient(&redis.Options{
			Addr: s.Addr(),
		})
		if _, err := c.Ping().Result(); err != nil {
			t.Errorf("Unable to connect to miniredis server: %s", err)
		}

		t.Run("Simple data structures", func(t *testing.T) {
			t.Run("Retrieving sample data", func(t *testing.T) {
				for _, tt := range redisSampleKV {
					redisGET(c, tt.key, tt.value, t)
				}

				for _, tt := range redisSampleHash {
					r, err := c.HGetAll(tt.key).Result()
					if err != nil {
						t.Errorf("Unable to HGetAll key %s: %s", tt.key, err)
					}
					for f, v := range tt.fv {
						if r[f] != v {
							t.Errorf("HGetAll %s, expected: %s => %s, got: %s => %s", tt.key, f, v, f, r[f])
						}
					}
				}
			})

			t.Run("Deleting sample data", func(t *testing.T) {
				t.Run("Deleting keys", func(t *testing.T) {
					for _, tt := range redisSampleKV {
						redisDEL(c, tt.key, 1, t)
					}

					for _, tt := range redisSampleHash {
						redisDEL(c, tt.key, 1, t)
					}
				})
				t.Run("Testing for existence", func(t *testing.T) {
					for _, tt := range redisSampleKV {
						redisEXISTS(c, tt.key, 0, t)
					}

					for _, tt := range redisSampleHash {
						redisEXISTS(c, tt.key, 0, t)
					}
				})
			})

			t.Run("Recreating sample data", func(t *testing.T) {
				t.Run("Setting data", func(t *testing.T) {
					for _, tt := range redisSampleKV {
						redisSET(c, tt.key, tt.value, t)
					}

					for _, tt := range redisSampleHash {
						for f, v := range tt.fv {
							redisHSET(c, tt.key, f, v, true, t)
						}
					}
				})

				t.Run("Testing for existence", func(t *testing.T) {
					for _, tt := range redisSampleKV {
						redisEXISTS(c, tt.key, 1, t)
					}

					for _, tt := range redisSampleHash {
						redisEXISTS(c, tt.key, 1, t)
					}
				})

				t.Run("Getting data", func(t *testing.T) {
					for _, tt := range redisSampleKV {
						redisGET(c, tt.key, tt.value, t)
					}

					for _, tt := range redisSampleHash {
						redisHGETALL(c, tt.key, tt.fv, t)
					}
				})
			})
		})
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
			s.serveHTTP(w, req)
			if w.Code != tt.wantStatus {
				t.Errorf("Wrong status code on request %s %s. Got: %d, want: %d", tt.method, tt.path, w.Code, tt.wantStatus)
			}
		}
	})
}
