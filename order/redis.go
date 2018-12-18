package order

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	ot "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
)

func appendNamespace(id string) string {
	return fmt.Sprintf("%s:%s", orderKeyNamespace, id)
}

func removeNamespace(key string) string {
	str := strings.Split(key, ":")[1:]
	return strings.Join(str, "")
}

// RedisScanOrders retrieves the IDs (keys) of all orders.
func (s *Server) RedisScanOrders(ctx context.Context) ([]int64, error) {
	span, _ := ot.StartSpanFromContext(ctx, "RedisScanOrders")
	defer span.Finish()

	var cursor uint64
	var keys []int64
	var err error

	for {
		var ks []string
		ks, cursor, err = s.redis.Scan(cursor, fmt.Sprintf("%s:*", orderKeyNamespace), 10).Result()
		if err != nil {
			return nil, err
		}

		for _, k := range ks {
			id, err := strconv.ParseInt(removeNamespace(k), 10, 64)
			if err != nil {
				return nil, err
			}
			keys = append(keys, id)
		}

		if cursor == 0 {
			break
		}
	}

	return keys, err
}

// RedisGetNextOrderID retrieves increments the order ID counter in redis and returns it.
func (s *Server) RedisGetNextOrderID(ctx context.Context) (int64, error) {
	span, _ := ot.StartSpanFromContext(ctx, "RedisGetNextOrderID")
	defer span.Finish()

	r, err := s.redis.Incr(nextIDKey).Result()
	if err != nil {
		return -1, err
	}
	return r, nil
}

// RedisSetOrder creates or updates a new order in Redis.
func (s *Server) RedisSetOrder(ctx context.Context, o *Order) error {
	span, _ := ot.StartSpanFromContext(ctx, "RedisScanOrders")
	defer span.Finish()

	if o.Items == nil {
		return errors.Errorf("order needs items, is %#v", o.Items)
	}

	id, items := o.MarshalRedis()
	key := appendNamespace(id)
	for k, v := range items {
		if err := s.redis.HSet(key, k, v).Err(); err != nil {
			return err
		}
	}

	return nil
}

// RedisGetOrder retrieves a single order from Redis.
func (s *Server) RedisGetOrder(ctx context.Context, id int64) (*Order, error) {
	span, _ := ot.StartSpanFromContext(ctx, "RedisGetOrder")
	defer span.Finish()

	key := strconv.FormatInt(id, 10)
	key = appendNamespace(key)
	r, err := s.redis.HGetAll(key).Result()
	if err != nil {
		return nil, err
	}

	if len(r) == 0 {
		return nil, nil
	}

	items := make(map[string]int)
	for k, v := range r {
		qty, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		items[k] = qty
	}

	o := &Order{}
	err = UnmarshalRedis(removeNamespace(key), items, o)
	return o, err
}
