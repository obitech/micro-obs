package item

// ScanKeys retrieves all keys from a redis instance.
// This uses the SCAN command so it's save to use on large database & in production.
func (s *Server) ScanKeys() ([]string, error) {
	var cursor uint64
	var keys []string
	var err error

	for {
		var k []string
		k, cursor, err = s.redis.Scan(cursor, "", 10).Result()
		keys = append(keys, k...)
		if err != nil {
			return nil, err
		}
		if cursor == 0 {
			break
		}
	}
	return keys, err
}

// GetItem retrieves an Item from Redis.
// func (s *Server) GetItem(k string) (*Item, error) {
// 	r, err := s.redis.HGetAll(k).Result()
// 	if err != nil {
// 		return nil, err
// 	}

// 	var item *Item
// 	return nil, nil
// }
