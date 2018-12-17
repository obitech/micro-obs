package item

import (
	"context"
	"reflect"
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

func TestItemRedis(t *testing.T) {
	// Setup miniredis
	_, mr := helperPrepareMiniredis(t)
	defer mr.Close()

	addr := strings.Join([]string{"redis://", mr.Addr()}, "")
	s, err := NewServer(
		SetRedisAddress(addr),
	)
	if err != nil {
		t.Errorf("unable to create server: %s", err)
	}

	var sampleKeys []string
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

			if err := s.RedisSetItem(context.Background(), i); err != nil {
				t.Error(err)
			}
			sampleKeys = append(sampleKeys, i.ID)
		}
	})

	t.Run("ScanKeys function", func(t *testing.T) {
		keys, err := s.RedisScanKeys(context.Background())
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
			i, err := s.RedisGetItem(context.Background(), k)
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

	t.Run("DelItems function", func(t *testing.T) {
		var items = make([]*Item, len(sampleItems))
		for i, v := range sampleItems {
			items[i], _ = NewItem(v.name, v.desc, v.qty)
		}

		err := s.RedisDelItems(context.Background(), items)
		if err != nil {
			t.Errorf("unable to delete itemes: %s", err)
		}
	})

	t.Run("Recreting items", func(t *testing.T) {
		for _, tt := range sampleItems {
			i, err := NewItem(tt.name, tt.desc, tt.qty)
			if err != nil {
				t.Errorf("unable to create new item: %s", err)
			}

			if err := s.RedisSetItem(context.Background(), i); err != nil {
				t.Error(err)
			}
			sampleKeys = append(sampleKeys, i.ID)
		}
	})

	t.Run("DelItem function", func(t *testing.T) {
		for _, k := range sampleKeys {
			err := s.RedisDelItem(context.Background(), k)
			if err != nil {
				t.Errorf("unable to delete key %#v: %s", k, err)
			}
		}
	})
}
