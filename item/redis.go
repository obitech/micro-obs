package item

import (
	"github.com/pkg/errors"
)

// ScanKeys retrieves all keys from a redis instance.
// This uses the SCAN command so it's save to use on large database & in production.
func (s *Server) ScanKeys() ([]string, error) {
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

// GetItem retrieves an Item from Redis.
func (s *Server) GetItem(k string) (*Item, error) {
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

// SetItem sets an Item as a hash in Redis.
func (s *Server) SetItem(i *Item) error {
	k, fv := i.MarshalRedis()
	for f, v := range fv {
		_, err := s.redis.HSet(k, f, v).Result()
		if err != nil {
			return errors.Errorf("unable to HSET %s %s %s", k, f, v)
		}
	}

	return nil
}

// DelItems deletes one or more Items from Redis.
func (s *Server) DelItems(items []*Item) error {
	var keys = make([]string, len(items))
	for i, v := range items {
		keys[i] = v.ID
	}

	err := s.redis.Del(keys...).Err()
	return err
}

// DelItem deletes a single Item by ID.
func (s *Server) DelItem(id string) error {
	return s.redis.Del(id).Err()
}
