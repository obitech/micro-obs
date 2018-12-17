package order

import (
	_ "context"
	_ "reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis"
)

func helperPrepareMiniredis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	mr, _ := miniredis.Run()
	c := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return c, mr
}

func TestOrderRedis(t *testing.T) {
	// Init miniredis
	_, mr := helperPrepareMiniredis(t)
	defer mr.Close()

	s, err := NewServer(
		SetRedisAddress(strings.Join([]string{"redis://", mr.Addr()}, "")),
	)
	if err != nil {
		t.Errorf("unable to create server: %s", err)
	}

	t.Run("Pinging miniredis with Server", func(t *testing.T) {
		if _, err := s.redis.Ping().Result(); err != nil {
			t.Errorf("unable to ping miniredis: %s", err)
		}
	})

	var wantID int64 = 1
	t.Run("RedisGetNextOrderID should be 1 when called first", func(t *testing.T) {
		id, err := s.RedisGetNextOrderID()
		if err != nil {
			t.Errorf("unable to INCR %s: %s", nextIDKey, err)
		}
		if id != wantID {
			t.Errorf("id mismatch, got: %d, want: %d", id, wantID)
		}

		r, err := s.redis.Get(nextIDKey).Result()
		if err != nil {
			t.Errorf("unable to GET %s: %s", nextIDKey, err)
		}
		verify, err := strconv.ParseInt(r, 10, 64)
		if err != nil {
			t.Errorf("unable to convert %s to int64: %s", r, err)
		}
		if verify != wantID {
			t.Errorf("id mismatch, got: %d, want: %d", id, wantID)
		}
	})

	t.Run("Incremnting nextID", func(t *testing.T) {
		wantID = 2
		id, err := s.RedisGetNextOrderID()
		if err != nil {
			t.Errorf("unable to INCR %s: %s", nextIDKey, err)
		}
		if id != wantID {
			t.Errorf("id mismatch, got: %d, want: %d", id, wantID)
		}

		var incBy int64 = 6
		var i int64 = 0
		for i = 0; i < incBy; i++ {
			id, err = s.RedisGetNextOrderID()
			if err != nil {
				t.Errorf("unable to INCR %s: %s", nextIDKey, err)
			}
		}
		if id != wantID+incBy {
			if id != wantID {
				t.Errorf("id mismatch, got: %d, want: %d", id, wantID+incBy)
			}
		}
	})
}
