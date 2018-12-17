package order

import (
	_ "context"

	_ "github.com/opentracing/opentracing-go"
	_ "github.com/pkg/errors"
)

// RedisGetNextOrderID retrieves increments the order ID counter in redis and returns it.
func (s *Server) RedisGetNextOrderID() (int64, error) {
	r, err := s.redis.Incr(nextIDKey).Result()
	if err != nil {
		return -1, err
	}
	return r, nil
}
