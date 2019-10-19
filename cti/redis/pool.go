package redis

import (
	"log"
	"os"
	"time"

	"github.com/garyburd/redigo/redis"
)

var (
	// Pool redis
	Pool *redis.Pool
)

func init() {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "redis:6379"
	}
	Pool = newPool(redisHost)
	Ping()
}

func newPool(server string) *redis.Pool {

	return &redis.Pool{

		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,

		Dial: func() (redis.Conn, error) {
			log.Printf("Connecting to Redis: %s...\n", server)
			connectTicker := time.Tick(time.Second * 2)
			var err error
			var c redis.Conn
			max := 5
		LOOP:
			for {
				select {
				case <-connectTicker:
					c, err = redis.Dial("tcp", server)
					if err == nil {
						break LOOP
					}
					if max == 0 {
						log.Fatalln("Failed to connect to Redis")
					}
					max--
				}
			}
			if err != nil {
				return nil, err
			}
			return c, err
		},

		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}
