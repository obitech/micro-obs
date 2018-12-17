package order

import (
	_ "context"
	_ "reflect"
	_ "strings"
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

// TODO: Test redisInitNextID

// TODO: Test RedisGetNextOrderID
