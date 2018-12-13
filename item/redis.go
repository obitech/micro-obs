package item

import "fmt"

// GetAllKeys retrieves all keys from a redis instance.
// This uses the SCAN command so it's save to use on large database & in production.
func (s *Server) GetAllKeys() ([]string, error) {
	var cursor uint64
	var keys []string
	var err error

	for {
		var k []string
		k, cursor, err = s.redis.Scan(cursor, "", 10).Result()
		fmt.Println(k)
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
