package order

import (
	_ "context"

	_ "github.com/opentracing/opentracing-go"
	_ "github.com/pkg/errors"
)

const (
	nextIDKey = "nextID"
)

// RedisGetNextOrderID retrieves increments the order ID counter in redis and returns it.
// TODO: trace this func
func (s *Server) RedisGetNextOrderID() (int64, error) {
	r, err := s.redis.Incr(nextIDKey).Result()
	if err != nil {
		return -1, err
	}
	return r, nil
}

// TODO: trace this func
func (s *Server) redisInitNextID() error {
	err := s.redis.SetNX(nextIDKey, 0, 0).Err()
	return err
}
