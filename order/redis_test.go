package order

import (
	"context"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis"
	"github.com/obitech/micro-obs/item"
)

func helperPrepareMiniredis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	mr, _ := miniredis.Run()
	c := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return c, mr
}

func helperMapsEqual(m1, m2 map[string]int) bool {
	if len(m1) != len(m2) {
		return false
	}

	for k, v := range m1 {
		if vv, prs := m2[k]; !prs || vv != v {
			return false
		}
	}

	return true
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
		id, err := s.RedisGetNextOrderID(context.Background())
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
		id, err := s.RedisGetNextOrderID(context.Background())
		if err != nil {
			t.Errorf("unable to INCR %s: %s", nextIDKey, err)
		}
		if id != wantID {
			t.Errorf("id mismatch, got: %d, want: %d", id, wantID)
		}

		var incBy int64 = 6
		var i int64
		for i = 0; i < incBy; i++ {
			id, err = s.RedisGetNextOrderID(context.Background())
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

	var uniqueSampleKeys []int64
	t.Run("Setting and getting new Order", func(t *testing.T) {

		allItems := []*Item{}
		for i, v := range sampleItems {
			item, err := item.NewItem(v.name, v.desc, v.qty)
			if err != nil {
				t.Errorf("unable to create item: %s", err)
			}

			oi, err := NewItem(item)
			if err != nil {
				t.Errorf("unable to create order item: %s", err)
			}
			allItems = append(allItems, oi)

			o, err := NewOrder(sampleOrderIDs[i], oi)
			uniqueOrders = append(uniqueOrders, o)
		}
		temp, err := NewOrder(int64(999999), allItems...)
		if err != nil {
			t.Errorf("unable to create order: %s", err)
		}
		uniqueOrders = append(uniqueOrders, temp)

		for _, o := range uniqueOrders {
			if err := s.RedisSetOrder(context.Background(), o); err != nil {
				t.Errorf("setting order failed: %s", err)
			}

			verify, err := s.RedisGetOrder(context.Background(), o.ID)
			if err != nil {
				t.Errorf("unable to get order: %s", err)
			}

			if o.ID != verify.ID {
				t.Errorf("%+v != %+v", o.ID, verify.ID)
			}

			if !reflect.DeepEqual(o.Items, verify.Items) {
				t.Errorf("%+v != %+v", o.Items, verify.Items)
			}
			uniqueSampleKeys = append(uniqueSampleKeys, o.ID)
		}
	})

	t.Run("Scan for all orders", func(t *testing.T) {
		keys, err := s.RedisScanOrders(context.Background())
		if err != nil {
			t.Errorf("unable to scan for orders: %s", err)
		}
		sort.Slice(keys, func(i, j int) bool {
			return keys[i] < keys[j]
		})
		sort.Slice(uniqueSampleKeys, func(i, j int) bool {
			return uniqueSampleKeys[i] < uniqueSampleKeys[j]
		})
		if !reflect.DeepEqual(keys, uniqueSampleKeys) {
			t.Errorf("%#v != %#v", keys, uniqueSampleKeys)
		}
	})
}
