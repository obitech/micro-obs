package item

import (
	"context"

	ot "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
)

// RedisScanKeys retrieves all keys from a redis instance.
// This uses the SCAN command so it's save to use on large database & in production.
func (s *Server) RedisScanKeys(ctx context.Context) ([]string, error) {
	span, _ := ot.StartSpanFromContext(ctx, "RedisScanKeys")
	defer span.Finish()

	var cursor uint64
	var keys []string
	var err error

	for {
		var k []string
		k, cursor, err = s.redis.Scan(cursor, "", 10).Result()
		if err != nil {
			return nil, err
		}
		keys = append(keys, k...)

		if cursor == 0 {
			break
		}
	}
	return keys, err
}

// RedisGetItem retrieves an Item from Redis.
func (s *Server) RedisGetItem(ctx context.Context, k string) (*Item, error) {
	span, _ := ot.StartSpanFromContext(ctx, "RedisGetItem")
	defer span.Finish()

	r, err := s.redis.HGetAll(k).Result()
	if err != nil {
		return nil, err
	}

	if len(r) == 0 {
		return nil, nil
	}

	var i = &Item{}
	err = UnmarshalRedis(k, r, i)

	return i, err
}

// RedisSetItem sets an Item as a hash in Redis.
func (s *Server) RedisSetItem(ctx context.Context, i *Item) error {
	span, _ := ot.StartSpanFromContext(ctx, "RedisSetItem")
	defer span.Finish()

	k, fv := i.MarshalRedis()
	for f, v := range fv {
		_, err := s.redis.HSet(k, f, v).Result()
		if err != nil {
			return errors.Errorf("unable to HSET %s %s %s", k, f, v)
		}
	}

	return nil
}

// RedisDelItems deletes one or more Items from Redis.
func (s *Server) RedisDelItems(ctx context.Context, items []*Item) error {
	span, _ := ot.StartSpanFromContext(ctx, "RedisDelItems")
	defer span.Finish()

	var keys = make([]string, len(items))
	for i, v := range items {
		keys[i] = v.ID
	}

	err := s.redis.Del(keys...).Err()
	return err
}

// RedisDelItem deletes a single Item by ID.
func (s *Server) RedisDelItem(ctx context.Context, id string) error {
	span, _ := ot.StartSpanFromContext(ctx, "RedisDelItems")
	defer span.Finish()

	return s.redis.Del(id).Err()
}
