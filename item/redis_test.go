package item

import (
	"reflect"
	"strings"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis"
)

var (
	sampleKVs = []struct {
		key   string
		value string
	}{
		{"foo", "bar"},
		{"hello", "world"},
		{"😍", "test"},
		{"test", "😍"},
		{"👾", "🙅"},
	}

	sampleHashes = []struct {
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

func helperPrepareMiniredis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	s, _ := miniredis.Run()
	c := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})
	return c, s
}

func helperRedisGET(c *redis.Client, k, want string, t *testing.T) {
	r, err := c.Get(k).Result()
	if err != nil {
		t.Errorf("unable to GET %#v: %#v", k, err)
	}
	if r != want {
		t.Errorf("GET %#v, expected: %#v, got: %#v", k, r, want)
	}
}

func helperRedisSET(c *redis.Client, k, v string, t *testing.T) {
	if err := c.Set(k, v, 0).Err(); err != nil {
		t.Errorf("unable to SET %#v -> %#v: %#v", k, v, err)
	}
}

func helperRedisHSET(c *redis.Client, k, f, v string, want bool, t *testing.T) {
	r, err := c.HSet(k, f, v).Result()
	if err != nil {
		t.Errorf("unable to HSET %#v %#v %#v: %#v", k, f, v, err)
	}
	if r != want {
		t.Errorf("HSET %#v %#v %#v, expected: %t, got: %t", k, f, v, want, r)
	}
}

func helperRedisDEL(c *redis.Client, k string, want int64, t *testing.T) {
	nr, err := c.Del(k).Result()
	if err != nil {
		t.Errorf("unable to DEL %#v: %#v", k, err)
	}
	if nr != want {
		t.Errorf("DEL %#v, expected: %d, got: %d", k, want, nr)
	}
}

func helperRedisEXISTS(c *redis.Client, k string, want int64, t *testing.T) {
	nr, err := c.Exists(k).Result()
	if err != nil {
		t.Errorf("unable to EXISTS %#v: %#v", k, err)
	}
	if nr != want {
		t.Errorf("EXISTS %#v, expected: %d, got: %d", k, want, nr)
	}
}

func helperRedisHGETALL(c *redis.Client, k string, want map[string]string, t *testing.T) {
	r, err := c.HGetAll(k).Result()
	if err != nil {
		t.Errorf("unable to HGETALL %#v: %#v", k, err)
	}
	for f, v := range want {
		if r[f] != v {
			t.Errorf("HGETALL %#v, expected: %#v => %#v, got: %#v => %#v", k, f, v, f, r[f])
		}
	}
}

func TestGoRedis(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Errorf("unable to create miniredis server: %#v", err)
	}
	defer s.Close()

	// Setting sample data
	for _, data := range sampleKVs {
		if err := s.Set(data.key, data.value); err != nil {
			t.Errorf("unable to set sample data: %#v", err)
		}
	}

	for _, data := range sampleHashes {
		for f, v := range data.fv {
			s.HSet(data.key, f, v)
		}
	}

	// Connecting
	c := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})
	if _, err := c.Ping().Result(); err != nil {
		t.Errorf("unable to connect to miniredis server: %#v", err)
	}

	t.Run("Retrieving sample data", func(t *testing.T) {
		for _, tt := range sampleKVs {
			helperRedisGET(c, tt.key, tt.value, t)
		}

		for _, tt := range sampleHashes {
			r, err := c.HGetAll(tt.key).Result()
			if err != nil {
				t.Errorf("unable to HGetAll key %#v: %#v", tt.key, err)
			}
			for f, v := range tt.fv {
				if r[f] != v {
					t.Errorf("HGetAll %#v, expected: %#v => %#v, got: %#v => %#v", tt.key, f, v, f, r[f])
				}
			}
		}
	})

	t.Run("Deleting sample data", func(t *testing.T) {
		t.Run("Deleting keys", func(t *testing.T) {
			for _, tt := range sampleKVs {
				helperRedisDEL(c, tt.key, 1, t)
			}

			for _, tt := range sampleHashes {
				helperRedisDEL(c, tt.key, 1, t)
			}
		})
		t.Run("Testing for existence", func(t *testing.T) {
			for _, tt := range sampleKVs {
				helperRedisEXISTS(c, tt.key, 0, t)
			}

			for _, tt := range sampleHashes {
				helperRedisEXISTS(c, tt.key, 0, t)
			}
		})
	})

	t.Run("Recreating sample data", func(t *testing.T) {
		t.Run("Setting data", func(t *testing.T) {
			for _, tt := range sampleKVs {
				helperRedisSET(c, tt.key, tt.value, t)
			}

			for _, tt := range sampleHashes {
				for f, v := range tt.fv {
					helperRedisHSET(c, tt.key, f, v, true, t)
				}
			}
		})

		t.Run("Testing for existence", func(t *testing.T) {
			for _, tt := range sampleKVs {
				helperRedisEXISTS(c, tt.key, 1, t)
			}

			for _, tt := range sampleHashes {
				helperRedisEXISTS(c, tt.key, 1, t)
			}
		})

		t.Run("Getting data", func(t *testing.T) {
			for _, tt := range sampleKVs {
				helperRedisGET(c, tt.key, tt.value, t)
			}

			for _, tt := range sampleHashes {
				helperRedisHGETALL(c, tt.key, tt.fv, t)
			}
		})
	})
}

func TestItemRedis(t *testing.T) {
	// Setup miniredis
	_, mr := helperPrepareMiniredis(t)
	defer mr.Close()

	var sampleKeys []string

	addr := strings.Join([]string{"redis://", mr.Addr()}, "")
	s, err := NewServer(
		SetRedisAddress(addr),
	)
	if err != nil {
		t.Errorf("unable to create server: %s", err)
	}

	t.Run("Pinging miniredis with Server", func(t *testing.T) {
		if _, err := s.redis.Ping().Result(); err != nil {
			t.Errorf("unable to ping miniredis: %s", err)
		}
	})

	t.Run("SetItems with sampleItems", func(t *testing.T) {
		for _, tt := range sampleItems {
			i, err := NewItem(tt.name, tt.desc, tt.qty)
			if err != nil {
				t.Errorf("unable to create new item: %s", err)
			}

			if err := s.SetItem(i); err != nil {
				t.Error(err)
			}
			sampleKeys = append(sampleKeys, i.ID)
		}
	})

	t.Run("ScanKeys function", func(t *testing.T) {
		keys, err := s.ScanKeys()
		if err != nil {
			t.Errorf("unable to SCAN redis for keys: %s", err)
		}

		if !reflect.DeepEqual(keys, sampleKeys) {
			t.Errorf("%#v != %#v", keys, sampleKeys)
		}
	})

	t.Run("GetItem function", func(t *testing.T) {
		var c int
		for _, k := range sampleKeys {
			i, err := s.GetItem(k)
			if err != nil {
				t.Errorf("unable to retrieve item with key %s: %s", k, err)
			}

			v, _ := NewItem(sampleItems[c].name, sampleItems[c].desc, sampleItems[c].qty)
			if !reflect.DeepEqual(i, v) {
				t.Errorf("%#v != %#v", i, v)
			}
			c++
		}
	})
}
